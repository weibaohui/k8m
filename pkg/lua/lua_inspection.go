package lua

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/kom/kom"
	lua "github.com/yuin/gopher-lua"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type Inspection struct {
	Cluster  string // 集群名称
	lua      *lua.LState
	Schedule *models.InspectionSchedule // 巡检计划ID
}

func NewLuaInspection(schedule *models.InspectionSchedule, cluster string) *Inspection {
	instance := &Inspection{
		Cluster:  cluster,
		Schedule: schedule,
		lua:      lua.NewState(),
	}
	instance.registerKubectlFunc()
	return instance
}

// 调用方法可参考pkg/models/lua_scripts_builtin.go中的示例
func (p *Inspection) registerKubectlFunc() {
	p.lua.SetGlobal("log", p.lua.NewFunction(logFunc))

	k := kom.Cluster(p.Cluster)
	if k == nil {
		klog.Errorf("巡检 集群【%s】，但是该集群未连接，巡检结果为失败", p.Cluster)
	}

	ud := p.lua.NewUserData()
	ud.Value = &Kubectl{k}
	p.lua.SetGlobal("kubectl", ud)

	// 设置元方法
	mt := p.lua.NewTypeMetatable("kubectl")
	p.lua.SetField(mt, "__index", p.lua.SetFuncs(p.lua.NewTable(), map[string]lua.LGFunction{
		"GVK":                 gvkFunc, // kubectl.GVK(group, version, kind) Kind首字母大写
		"WithLabelSelector":   withLabelSelectorFunc,
		"Name":                withNameFunc,
		"Namespace":           withNamespaceFunc,
		"AllNamespace":        withAllNamespaceFunc,
		"Cache":               withCacheFunc,
		"List":                listResource,
		"Doc":                 getDoc,
		"Get":                 getResource,
		"GetLogs":             getLogs,
		"GetPodResourceUsage": getPodResourceUsage,
	}))
	p.lua.SetMetatable(ud, mt)
}
func (p *Inspection) Start() []CheckResult {
	// 初始化 Lua 状态
	defer p.lua.Close()

	params := &dao.Params{
		PerPage: 10000000,
	}

	klog.V(6).Infof("p.Schedule.ScriptCodes: %v", utils.ToJSON(p.Schedule.ScriptCodes))

	// 执行所有 Lua 检查脚本并收集结果
	var results []CheckResult

	// 从数据库读取内置检查脚本
	script := models.InspectionLuaScript{}

	list, _, err := script.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("script_code in ?", strings.Split(p.Schedule.ScriptCodes, ","))
	})
	if err != nil {
		fmt.Println("无法从数据库读取检查脚本:", err)
		return results
	}
	for _, item := range list {
		result := p.runLuaCheck(item)
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
				if evt.Status != string(constants.LuaEventStatusNormal) {
					b, _ := json.Marshal(evt)
					fmt.Println(string(b))
				}
			}
		} else {
			fmt.Println("无结构化检测事件（未调用 check_event）")
		}
	}

	return results
}

// runLuaCheck 执行单个Lua脚本检查，支持超时控制
func (p *Inspection) runLuaCheck(item *models.InspectionLuaScript) CheckResult {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	var events []CheckEvent
	p.registerCheckEvent(&events, item)

	// 获取超时时间，如果未设置或为0，则使用默认60秒
	timeoutSeconds := item.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	timeout := time.Duration(timeoutSeconds) * time.Second

	start := time.Now()
	
	// 使用channel来处理超时
	type result struct {
		err error
	}
	resultChan := make(chan result, 1)
	
	// 在goroutine中执行Lua脚本
	go func() {
		err := p.lua.DoString(item.Script)
		resultChan <- result{err: err}
	}()

	var err error
	// 等待脚本执行完成或超时
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	select {
	case res := <-resultChan:
		err = res.err
	case <-ctx.Done():
		err = fmt.Errorf("脚本执行超时（%d秒）", timeoutSeconds)
		klog.Warningf("Lua脚本 [%s] 执行超时，超时时间: %d秒", item.Name, timeoutSeconds)
	}

	_ = w.Close()
	os.Stdout = origStdout

	_, _ = buf.ReadFrom(r)
	_ = r.Close() // 读取完立即关闭
	output := buf.String()
	end := time.Now()

	return CheckResult{
		Name:         item.Name,
		StartTime:    start,
		EndTime:      end,
		LuaRunOutput: output,
		LuaRunError:  err,
		Events:       events,
	}
}

// 注册 check_event 到 Lua，自动补充上下文
func (p *Inspection) registerCheckEvent(events *[]CheckEvent, item *models.InspectionLuaScript) {
	p.lua.SetGlobal("check_event", p.lua.NewFunction(func(L *lua.LState) int {
		status := L.CheckString(1)
		msg := L.CheckString(2)
		var extra map[string]any
		if L.GetTop() >= 3 {
			extraVal := L.CheckAny(3)
			if tbl, ok := extraVal.(*lua.LTable); ok {
				extra = lValueToGoValue(tbl).(map[string]any)
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
			Name:       name,
			Namespace:  namespace,
			Status:     status,
			Msg:        msg,
			Extra:      extra,
			ScriptName: item.Name,        // 检测脚本名称
			Kind:       item.Kind,        // 检查的资源类型
			CheckDesc:  item.Description, // 检查脚本内容描述
		})
		return 0
	}))
}
