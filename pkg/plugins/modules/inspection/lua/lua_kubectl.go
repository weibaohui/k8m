package lua

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/kom/kom"
	lua "github.com/yuin/gopher-lua"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type Kubectl struct {
	k *kom.Kubectl
}

// 实现 kubectl:GVK(group, version, kind) 方法
func gvkFunc(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	// 获取 GVK 相关参数
	group := L.CheckString(2)
	version := L.CheckString(3)
	kind := L.CheckString(4)
	klog.V(6).Infof("执行GVK查询: %s/%s/%s", group, version, kind)
	// 确保每次 GVK 查询，返回新的 LuaKubectl 实例链，避免嵌套调用时混乱

	ctx := utils.GetContextWithAdmin()
	newObj := &Kubectl{obj.k.GVK(group, version, kind).WithContext(ctx).RemoveManagedFields()}
	newUd := L.NewUserData()
	newUd.Value = newObj
	L.SetMetatable(newUd, L.GetTypeMetatable("kubectl"))
	L.Push(newUd)
	L.Push(lua.LNil)

	return 2
}

// 实现 kubectl:WithLabelSelector(selector) 方法
func withLabelSelectorFunc(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	// 获取 labelSelector 参数
	selector := L.CheckString(2)
	if selector != "" {
		obj.k = obj.k.WithLabelSelector(selector)
	}
	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// 实现 kubectl:WithLabelSelector(selector) 方法
func withNameFunc(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	name := L.CheckString(2)
	if name != "" {
		obj.k = obj.k.Name(name)
	}
	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// 实现 kubectl:Namespace(ns) 方法
func withNamespaceFunc(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	name := L.CheckString(2)
	if name != "" {
		obj.k = obj.k.Namespace(name)
	}
	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// 实现 kubectl:Cache(t) 方法
// 该方法用于设置缓存时间，参数t为缓存时长（单位：秒）
func withCacheFunc(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	timeSeconds := L.CheckNumber(2)
	if timeSeconds > 0 {
		dur := time.Duration(int64(timeSeconds)) * time.Second
		obj.k = obj.k.WithCache(dur)
	}
	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// 实现 kubectl:AllNamespace() 方法
func withAllNamespaceFunc(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	obj.k = obj.k.AllNamespace()
	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

// 实现 kubectl:List() 方法
func listResource(L *lua.LState) int {
	klog.V(6).Infof("执行List查询")
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	// 查询资源
	var result []*unstructured.Unstructured
	err := obj.k.List(&result).Error
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// 转换为 Lua 表
	table := toLValue(L, result)
	// 返回查询结果
	L.Push(table)
	L.Push(lua.LNil)
	return 2
}

// 实现 kubectl:Get() 方法
// 用于获取单个资源，返回 Lua 表和错误信息
func getResource(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	// 查询单个资源
	var result *unstructured.Unstructured
	err := obj.k.Get(&result).Error
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// 转换为 Lua 表
	table := toLValue(L, result.Object)

	// 返回查询结果
	L.Push(table)
	L.Push(lua.LNil)
	return 2
}

// 实现 kubectl:Doc('spec.replicas') 方法
// 用于获取某个字段的解释，返回 Lua 表和错误信息
func getDoc(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	field := L.CheckString(2)
	if field != "" {
		obj.k = obj.k.DocField(field)
	}
	// 查询单个资源
	var result []byte
	err := obj.k.Doc(&result).Error
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// 转换为 Lua 表
	table := toLValue(L, string(result))

	// 返回查询结果
	L.Push(table)
	L.Push(lua.LNil)
	return 2
}

// 实现 kubectl:GetLogs({tailLines=100, container="xxx"}) 方法
// 返回日志文本和错误信息（若有）
// - local logs, err = kubectl.GVK("", "v1", "Pod").Namespace("default").Name("mypod").GetLogs({tailLines=200, container="app"})
// - if err ~= nil then print("error:", err) else print(logs) end
func getLogs(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	// 解析可选参数表
	var opt v1.PodLogOptions
	if L.GetTop() >= 2 {
		if tbl, ok := L.Get(2).(*lua.LTable); ok {
			// container 字段
			if v := tbl.RawGetString("container"); v.Type() == lua.LTString {
				opt.Container = v.String()
			} else if v := tbl.RawGetString("Container"); v.Type() == lua.LTString {
				opt.Container = v.String()
			}
			// tailLines 字段
			if v := tbl.RawGetString("tailLines"); v.Type() == lua.LTNumber {
				t := int64(lua.LVAsNumber(v))
				opt.TailLines = &t
			} else if v := tbl.RawGetString("TailLines"); v.Type() == lua.LTNumber {
				t := int64(lua.LVAsNumber(v))
				opt.TailLines = &t
			}
		}
	}

	var stream io.ReadCloser
	err := obj.k.Ctl().Pod().GetLogs(&stream, &opt).Error
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	if stream == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("empty log stream"))
		return 2
	}
	defer stream.Close()

	// 读取全部日志内容
	data, rerr := io.ReadAll(stream)
	if rerr != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(rerr.Error()))
		return 2
	}

	// 返回字符串
	L.Push(toLValue(L, string(data)))
	L.Push(lua.LNil)
	return 2
}

// 实现 kubectl:GetPodResourceUsage() 方法
// 用于获取Pod的资源使用情况，返回 Lua 表和错误信息
// 使用方式：local usage, err = kubectl:GVK("", "v1", "Pod"):Namespace("kube-system"):Name("coredns-ccb96694c-jprpf"):GetPodResourceUsage()
func getPodResourceUsage(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}

	// 调用kom库的ResourceUsage方法
	result, err := obj.k.Ctl().Pod().ResourceUsage(kom.DenominatorLimit)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// 转换为 Lua 表
	table := toLValue(L, result)

	// 返回查询结果
	L.Push(table)
	L.Push(lua.LNil)
	return 2
}

// parseLuaTime 将 Lua 传入的时间（数字/字符串）解析为 time.Time。
// - number：默认按 Unix 秒解析；当数值过大（>= 1e12）时按 Unix 毫秒解析
// - string：支持 RFC3339、RFC3339Nano、"2006-01-02 15:04:05"、纯数字（同 number）
func parseLuaTime(val lua.LValue) (time.Time, error) {
	switch v := val.(type) {
	case lua.LNumber:
		n := float64(v)
		if n >= 1e12 {
			return time.UnixMilli(int64(n)), nil
		}
		return time.Unix(int64(n), 0), nil
	case lua.LString:
		s := strings.TrimSpace(v.String())
		if s == "" {
			return time.Time{}, fmt.Errorf("时间为空")
		}
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			if i >= 1e12 {
				return time.UnixMilli(i), nil
			}
			return time.Unix(i, 0), nil
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
			if t, err := time.Parse(layout, s); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("无法解析时间: %s", s)
	default:
		return time.Time{}, fmt.Errorf("不支持的时间类型: %s", val.Type().String())
	}
}

// promResultToGoValue 将 Prometheus 查询结果转换为 Lua 可消费的数据结构。
// - resultType="string"：返回 res.AsString()
// - resultType="scalar"：返回 number
// - resultType="vector"：返回 [{metric=map,value=number,timestamp=string},...]
// - resultType="matrix"：返回 [{metric=map,samples=[{timestamp=string,value=number},...]},...]
func promResultToGoValue(res *kom.PromResult, resultType string) (any, error) {
	switch strings.ToLower(strings.TrimSpace(resultType)) {
	case "", "string":
		return res.AsString(), nil
	case "scalar":
		v, ok := res.AsScalar()
		if !ok {
			return nil, fmt.Errorf("查询结果不是标量")
		}
		return v, nil
	case "vector":
		samples := res.AsVector()
		out := make([]any, 0, len(samples))
		for _, s := range samples {
			out = append(out, map[string]any{
				"metric":    s.Metric,
				"value":     s.Value,
				"timestamp": s.Timestamp.Format(time.RFC3339),
			})
		}
		return out, nil
	case "matrix":
		series := res.AsMatrix()
		out := make([]any, 0, len(series))
		for _, se := range series {
			points := make([]any, 0, len(se.Samples))
			for _, p := range se.Samples {
				points = append(points, map[string]any{
					"timestamp": p.Timestamp.Format(time.RFC3339),
					"value":     p.Value,
				})
			}
			out = append(out, map[string]any{
				"metric":  se.Metric,
				"samples": points,
			})
		}
		return out, nil
	default:
		return res.AsString(), nil
	}
}

// promQuery 实现 kubectl:PromQuery(opts) 方法。
// 用于执行 Prometheus 瞬时查询，返回 Lua 值和错误信息。
// opts 支持字段：
// - address: string，外部 Prometheus 地址
// - namespace/service(or svcName): string，集群内 Prometheus Service 定位（优先使用 address，其次使用 namespace+service）
// - expr: string，PromQL 表达式
// - at: number|string，可选，指定瞬时时间点（不传则为当前时间）
// - timeoutSeconds: number，可选，查询超时（秒）
// - resultType: string，可选，string/scalar/vector/matrix（默认 string）
func promQuery(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}
	if obj.k == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("kubectl 未初始化"))
		return 2
	}
	tbl := L.CheckTable(2)

	exprVal := tbl.RawGetString("expr")
	if exprVal == lua.LNil {
		exprVal = tbl.RawGetString("Expr")
	}
	expr := strings.TrimSpace(exprVal.String())
	if expr == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("缺少 expr 参数"))
		return 2
	}

	address := strings.TrimSpace(tbl.RawGetString("address").String())
	if address == "" {
		address = strings.TrimSpace(tbl.RawGetString("Address").String())
	}
	namespace := strings.TrimSpace(tbl.RawGetString("namespace").String())
	if namespace == "" {
		namespace = strings.TrimSpace(tbl.RawGetString("Namespace").String())
	}
	if namespace == "" {
		namespace = strings.TrimSpace(tbl.RawGetString("inClusterNamespace").String())
	}
	service := strings.TrimSpace(tbl.RawGetString("service").String())
	if service == "" {
		service = strings.TrimSpace(tbl.RawGetString("Service").String())
	}
	if service == "" {
		service = strings.TrimSpace(tbl.RawGetString("svcName").String())
	}
	if service == "" {
		service = strings.TrimSpace(tbl.RawGetString("SvcName").String())
	}
	if service == "" {
		service = strings.TrimSpace(tbl.RawGetString("inClusterService").String())
	}

	timeoutSeconds := int64(0)
	if v := tbl.RawGetString("timeoutSeconds"); v.Type() == lua.LTNumber {
		timeoutSeconds = int64(lua.LVAsNumber(v))
	} else if v := tbl.RawGetString("TimeoutSeconds"); v.Type() == lua.LTNumber {
		timeoutSeconds = int64(lua.LVAsNumber(v))
	}

	resultType := strings.TrimSpace(tbl.RawGetString("resultType").String())
	if resultType == "" {
		resultType = strings.TrimSpace(tbl.RawGetString("ResultType").String())
	}
	if resultType == "" {
		resultType = "string"
	}

	ctx := utils.GetContextWithAdmin()

	client := obj.k.WithContext(ctx).Prometheus()
	var query *kom.PromQuery
	if address != "" {
		klog.V(6).Infof("执行 Prometheus 瞬时查询，使用外部地址：%s", address)
		query = client.WithAddress(address).Expr(expr)
	} else {
		klog.V(6).Infof("执行 Prometheus 瞬时查询，使用集群内服务：%s/%s", namespace, service)
		query = client.WithInClusterEndpoint(namespace, service).Expr(expr)
	}
	if timeoutSeconds > 0 {
		query = query.WithTimeout(time.Duration(timeoutSeconds) * time.Second)
	}

	var (
		res *kom.PromResult
		err error
	)
	if v := tbl.RawGetString("at"); v != lua.LNil {
		at, perr := parseLuaTime(v)
		if perr != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(perr.Error()))
			return 2
		}
		klog.V(6).Infof("执行 Prometheus 瞬时查询，指定时间点：%s", at.Format(time.RFC3339))
		res, err = query.QueryAt(at)
	} else if v := tbl.RawGetString("At"); v != lua.LNil {
		at, perr := parseLuaTime(v)
		if perr != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(perr.Error()))
			return 2
		}
		klog.V(6).Infof("执行 Prometheus 瞬时查询，指定时间点：%s", at.Format(time.RFC3339))
		res, err = query.QueryAt(at)
	} else {
		klog.V(6).Infof("执行 Prometheus 瞬时查询，当前时间点")
		res, err = query.Query()
	}
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	out, terr := promResultToGoValue(res, resultType)
	if terr != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(terr.Error()))
		return 2
	}

	L.Push(toLValue(L, out))
	L.Push(lua.LNil)
	return 2
}

// promQueryRange 实现 kubectl:PromQueryRange(opts) 方法。
// 用于执行 Prometheus 区间查询，返回 Lua 值和错误信息。
// opts 支持字段：
// - address: string，外部 Prometheus 地址
// - namespace/service(or svcName): string，集群内 Prometheus Service 定位（优先使用 address，其次使用 namespace+service）
// - expr: string，PromQL 表达式
// - start/end: number|string，必填，时间范围
// - stepSeconds/step: number|string，可选，步长（秒或 "1m" 这类 duration，默认 60 秒）
// - timeoutSeconds: number，可选，查询超时（秒）
// - resultType: string，可选，string/matrix（默认 string）
func promQueryRange(L *lua.LState) int {
	ud := L.CheckUserData(1)
	obj, ok := ud.Value.(*Kubectl)
	if !ok {
		L.ArgError(1, "expected kubectl")
		return 0
	}
	if obj.k == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("kubectl 未初始化"))
		return 2
	}
	tbl := L.CheckTable(2)

	exprVal := tbl.RawGetString("expr")
	if exprVal == lua.LNil {
		exprVal = tbl.RawGetString("Expr")
	}
	expr := strings.TrimSpace(exprVal.String())
	if expr == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("缺少 expr 参数"))
		return 2
	}

	startVal := tbl.RawGetString("start")
	if startVal == lua.LNil {
		startVal = tbl.RawGetString("Start")
	}
	endVal := tbl.RawGetString("end")
	if endVal == lua.LNil {
		endVal = tbl.RawGetString("End")
	}
	if startVal == lua.LNil || endVal == lua.LNil {
		L.Push(lua.LNil)
		L.Push(lua.LString("缺少 start/end 参数"))
		return 2
	}
	start, err := parseLuaTime(startVal)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	end, err := parseLuaTime(endVal)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	step := time.Minute
	if v := tbl.RawGetString("stepSeconds"); v.Type() == lua.LTNumber {
		sec := int64(lua.LVAsNumber(v))
		if sec > 0 {
			step = time.Duration(sec) * time.Second
		}
	} else if v := tbl.RawGetString("StepSeconds"); v.Type() == lua.LTNumber {
		sec := int64(lua.LVAsNumber(v))
		if sec > 0 {
			step = time.Duration(sec) * time.Second
		}
	} else if v := tbl.RawGetString("step"); v.Type() == lua.LTNumber {
		sec := int64(lua.LVAsNumber(v))
		if sec > 0 {
			step = time.Duration(sec) * time.Second
		}
	} else if v := tbl.RawGetString("Step"); v.Type() == lua.LTNumber {
		sec := int64(lua.LVAsNumber(v))
		if sec > 0 {
			step = time.Duration(sec) * time.Second
		}
	} else if v := tbl.RawGetString("step"); v.Type() == lua.LTString {
		if d, derr := time.ParseDuration(strings.TrimSpace(v.String())); derr == nil && d > 0 {
			step = d
		}
	} else if v := tbl.RawGetString("Step"); v.Type() == lua.LTString {
		if d, derr := time.ParseDuration(strings.TrimSpace(v.String())); derr == nil && d > 0 {
			step = d
		}
	}

	timeoutSeconds := int64(0)
	if v := tbl.RawGetString("timeoutSeconds"); v.Type() == lua.LTNumber {
		timeoutSeconds = int64(lua.LVAsNumber(v))
	} else if v := tbl.RawGetString("TimeoutSeconds"); v.Type() == lua.LTNumber {
		timeoutSeconds = int64(lua.LVAsNumber(v))
	}

	resultType := strings.TrimSpace(tbl.RawGetString("resultType").String())
	if resultType == "" {
		resultType = strings.TrimSpace(tbl.RawGetString("ResultType").String())
	}
	if resultType == "" {
		resultType = "string"
	}

	address := strings.TrimSpace(tbl.RawGetString("address").String())
	if address == "" {
		address = strings.TrimSpace(tbl.RawGetString("Address").String())
	}
	namespace := strings.TrimSpace(tbl.RawGetString("namespace").String())
	if namespace == "" {
		namespace = strings.TrimSpace(tbl.RawGetString("Namespace").String())
	}
	if namespace == "" {
		namespace = strings.TrimSpace(tbl.RawGetString("inClusterNamespace").String())
	}
	service := strings.TrimSpace(tbl.RawGetString("service").String())
	if service == "" {
		service = strings.TrimSpace(tbl.RawGetString("Service").String())
	}
	if service == "" {
		service = strings.TrimSpace(tbl.RawGetString("svcName").String())
	}
	if service == "" {
		service = strings.TrimSpace(tbl.RawGetString("SvcName").String())
	}
	if service == "" {
		service = strings.TrimSpace(tbl.RawGetString("inClusterService").String())
	}

	ctx := utils.GetContextWithAdmin()

	client := obj.k.WithContext(ctx).Prometheus()
	var query *kom.PromQuery
	if address != "" {
		klog.V(6).Infof("执行 Prometheus 区间查询，使用外部地址：%s", address)
		query = client.WithAddress(address).Expr(expr)
	} else {
		klog.V(6).Infof("执行 Prometheus 区间查询，使用集群内服务：%s/%s", namespace, service)
		query = client.WithInClusterEndpoint(namespace, service).Expr(expr)
	}
	if timeoutSeconds > 0 {
		query = query.WithTimeout(time.Duration(timeoutSeconds) * time.Second)
	}

	klog.V(6).Infof("执行 Prometheus 区间查询，start=%s，end=%s，step=%s", start.Format(time.RFC3339), end.Format(time.RFC3339), step.String())
	res, err := query.QueryRange(start, end, step)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	out, terr := promResultToGoValue(res, resultType)
	if terr != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(terr.Error()))
		return 2
	}

	L.Push(toLValue(L, out))
	L.Push(lua.LNil)
	return 2
}
