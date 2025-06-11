package lua

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/kom/kom"
	lua "github.com/yuin/gopher-lua"
)

type Inspection struct {
	Cluster string // 集群名称
	lua     *lua.LState
}

func NewLuaInspection(cluster string) *Inspection {
	instance := &Inspection{
		Cluster: cluster,
		lua:     lua.NewState(),
	}
	instance.registerKubectlFunc()
	return instance
}
func (p *Inspection) registerKubectlFunc() {
	p.lua.SetGlobal("log", p.lua.NewFunction(logFunc))

	k := kom.DefaultCluster()

	ud := p.lua.NewUserData()
	ud.Value = &LuaKubectl{k}
	p.lua.SetGlobal("kubectl", ud)

	// 设置元方法
	mt := p.lua.NewTypeMetatable("kubectl")
	p.lua.SetField(mt, "__index", p.lua.SetFuncs(p.lua.NewTable(), map[string]lua.LGFunction{
		"GVK":               gvkFunc, // kubectl.GVK(group, version, kind) Kind首字母大写
		"WithLabelSelector": withLabelSelectorFunc,
		"Name":              withNameFunc,
		"Namespace":         withNamespaceFunc,
		"AllNamespace":      withAllNamespaceFunc,
		"Cache":             withCacheFunc,
		"List":              listResource,
		"Get":               getResource,
	}))
	p.lua.SetMetatable(ud, mt)
}
func (p *Inspection) Start() {
	// 初始化 Lua 状态
	defer p.lua.Close()

	// 执行所有 Lua 检查脚本并收集结果
	var results []CheckResult
	for _, check := range luaScripts {
		result := p.runLuaCheck(check)
		results = append(results, result)
	}

	// 打印所有检测结果
	for _, res := range results {
		fmt.Printf("\n--- 检查: %s ---\n", res.Name)
		fmt.Printf("开始时间: %s\n结束时间: %s\n", res.StartTime.Format(time.RFC3339), res.EndTime.Format(time.RFC3339))
		if res.LuaRunError != nil {
			fmt.Printf("错误: %v\n", res.LuaRunError)
		}
		fmt.Printf("输出:\n%s\n", res.LuaRunOutput)
		if len(res.Events) > 0 {
			fmt.Println("结构化检测失败事件:")
			for _, evt := range res.Events {
				if evt.Status != "正常" {
					b, _ := json.Marshal(evt)
					fmt.Println(string(b))
				}
			}
		} else {
			fmt.Println("无结构化检测事件（未调用 check_event）")
		}
	}
}

func (p *Inspection) runLuaCheck(check models.LuaScript) CheckResult {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var events []CheckEvent
	p.registerCheckEvent(&events, check)

	start := time.Now()
	err := p.lua.DoString(check.Script)
	_ = w.Close()
	os.Stdout = origStdout

	_, _ = buf.ReadFrom(r)
	output := buf.String()
	end := time.Now()

	return CheckResult{
		Name:         check.Name,
		StartTime:    start,
		EndTime:      end,
		LuaRunOutput: output,
		LuaRunError:  err,
		Events:       events,
	}
}

// 注册 check_event 到 Lua，自动补充上下文
func (p *Inspection) registerCheckEvent(events *[]CheckEvent, check models.LuaScript) {
	p.lua.SetGlobal("check_event", p.lua.NewFunction(func(L *lua.LState) int {
		status := L.CheckString(1)
		msg := L.CheckString(2)
		var extra map[string]interface{}
		if L.GetTop() >= 3 {
			extraVal := L.CheckAny(3)
			if tbl, ok := extraVal.(*lua.LTable); ok {
				extra = lValueToGoValue(tbl).(map[string]interface{})
			}
		}
		var name, namespace string
		if v, ok := extra["name"]; ok {
			name, _ = v.(string)
		}
		if v, ok := extra["namespace"]; ok {
			namespace, _ = v.(string)
		}
		*events = append(*events, CheckEvent{
			Name:            name,
			Namespace:       namespace,
			Status:          status,
			Msg:             msg,
			Extra:           extra,
			CheckScriptName: check.Name,        // 检测脚本名称
			Kind:            check.Kind,        // 检查的资源类型
			CheckDesc:       check.Description, // 检查脚本内容描述
		})
		return 0
	}))
}
