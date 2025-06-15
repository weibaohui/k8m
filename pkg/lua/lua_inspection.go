package lua

import (
	"bytes"
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
func (p *Inspection) registerKubectlFunc() {
	p.lua.SetGlobal("log", p.lua.NewFunction(logFunc))

	k := kom.DefaultCluster()

	ud := p.lua.NewUserData()
	ud.Value = &Kubectl{k}
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
		"Doc":               getDoc,
		"Get":               getResource,
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

func (p *Inspection) runLuaCheck(item *models.InspectionLuaScript) CheckResult {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	var events []CheckEvent
	p.registerCheckEvent(&events, item)

	start := time.Now()
	err := p.lua.DoString(item.Script)
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
