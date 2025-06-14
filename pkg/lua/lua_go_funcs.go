package lua

import (
	"encoding/json"
	"reflect"

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

// Lua LValue 转 Go interface{}，递归处理 table
func lValueToGoValue(val lua.LValue) interface{} {
	switch v := val.(type) {
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		// 判断是数组还是 map
		var arr []interface{}
		mp := map[string]interface{}{}
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
func toLValue(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case []*unstructured.Unstructured:
		items := make([]interface{}, len(val))
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
	case []interface{}:
		tbl := L.NewTable()
		for _, item := range val {
			tbl.Append(toLValue(L, item))
		}
		return tbl
	case map[string]interface{}:
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
