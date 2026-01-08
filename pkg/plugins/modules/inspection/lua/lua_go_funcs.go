package lua

import (
	"encoding/json"
	"reflect"

	"github.com/weibaohui/kom/kom"
	lua "github.com/yuin/gopher-lua"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// 增加 log 方法，支持 Lua 调用 log(obj)，自动使用 json.Marshal 格式化输出
func logFunc(L *lua.LState) int {
	value := L.CheckAny(1)
	goValue := lValueToGoValue(value)
	data, err := json.MarshalIndent(goValue, "", "  ")
	if err != nil {
		klog.V(6).Infof("[log] json.Marshal error:%v", err)
	} else {
		klog.V(6).Infof("[log] %s", string(data))
	}
	return 0
}

// Lua LValue 转 Go any，递归处理 table
func lValueToGoValue(val lua.LValue) any {
	switch v := val.(type) {
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		// 判断是数组还是 map
		var arr []any
		mp := map[string]any{}
		isArray := true
		maxIdx := 0
		v.ForEach(func(key, value lua.LValue) {
			if key.Type() == lua.LTNumber {
				idx := int(lua.LVAsNumber(key))
				if idx > maxIdx {
					maxIdx = idx
				}
				for len(arr) < idx {
					arr = append(arr, nil)
				}
				arr[idx-1] = lValueToGoValue(value)
			} else {
				isArray = false
				mp[lua.LVAsString(key)] = lValueToGoValue(value)
			}
		})
		if isArray && maxIdx > 0 {
			return arr
		}
		return mp
	default:
		klog.V(6).Infof("[lValueToGoValue] unknown type: %v", reflect.TypeOf(val))
		return nil
	}
}

// Go -> Lua 转换
func toLValue(L *lua.LState, v any) lua.LValue {
	switch val := v.(type) {
	case *kom.ResourceUsageResult:
		if val == nil {
			return lua.LNil
		}
		tbl := L.NewTable()

		// 转换 Requests
		if val.Requests != nil {
			requestsTbl := L.NewTable()
			for name, quantity := range val.Requests {
				requestsTbl.RawSetString(string(name), lua.LNumber(quantity.AsApproximateFloat64()))
			}
			tbl.RawSetString("requests", requestsTbl)
		}

		// 转换 Limits
		if val.Limits != nil {
			limitsTbl := L.NewTable()
			for name, quantity := range val.Limits {
				limitsTbl.RawSetString(string(name), lua.LNumber(quantity.AsApproximateFloat64()))
			}
			tbl.RawSetString("limits", limitsTbl)
		}

		// 转换 Realtime
		if val.Realtime != nil {
			realtimeTbl := L.NewTable()
			for name, quantity := range val.Realtime {
				realtimeTbl.RawSetString(string(name), lua.LNumber(quantity.AsApproximateFloat64()))
			}
			tbl.RawSetString("realtime", realtimeTbl)
		}

		// 转换 Allocatable
		if val.Allocatable != nil {
			allocatableTbl := L.NewTable()
			for name, quantity := range val.Allocatable {
				allocatableTbl.RawSetString(string(name), lua.LNumber(quantity.AsApproximateFloat64()))
			}
			tbl.RawSetString("allocatable", allocatableTbl)
		}

		// 转换 UsageFractions
		if val.UsageFractions != nil {
			usageFractionsTbl := L.NewTable()
			for name, fraction := range val.UsageFractions {
				fractionTbl := L.NewTable()
				fractionTbl.RawSetString("requestFraction", lua.LString(fraction.RequestFraction))
				fractionTbl.RawSetString("limitFraction", lua.LString(fraction.LimitFraction))
				fractionTbl.RawSetString("realtimeFraction", lua.LString(fraction.RealtimeFraction))
				usageFractionsTbl.RawSetString(string(name), fractionTbl)
			}
			tbl.RawSetString("usageFractions", usageFractionsTbl)
		}

		return tbl
	case []*unstructured.Unstructured:
		items := make([]any, len(val))
		for i, item := range val {
			items[i] = item.Object
		}
		return toLValue(L, items)
	case string:
		return lua.LString(val)
	case float64:
		return lua.LNumber(val)
	case float32:
		return lua.LNumber(val)
	case int:
		return lua.LNumber(val)
	case int64:
		return lua.LNumber(val)
	case int32:
		return lua.LNumber(val)
	case int16:
		return lua.LNumber(val)
	case int8:
		return lua.LNumber(val)
	case uint:
		return lua.LNumber(val)
	case uint64:
		return lua.LNumber(val)
	case uint32:
		return lua.LNumber(val)
	case uint16:
		return lua.LNumber(val)
	case uint8:
		return lua.LNumber(val)
	case bool:
		return lua.LBool(val)
	case []any:
		tbl := L.NewTable()
		for _, item := range val {
			tbl.Append(toLValue(L, item))
		}
		return tbl
	case map[string]any:
		tbl := L.NewTable()
		for k, v := range val {
			tbl.RawSetString(k, toLValue(L, v))
		}
		return tbl
	case nil:
		return lua.LNil
	default:
		klog.V(6).Infof("[toLValue] unknown type: %v", reflect.TypeOf(v))
		return lua.LNil
	}
}
