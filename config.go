package minion

import (
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/lua"
)

type config struct {
	Type string
	Name string
	Data map[string]any
}

func toArr(arr []lua.LValue, x func(i interface{})) {
	var data []interface{}
	for _, item := range arr {
		convert(item, func(i interface{}) {
			data = append(data, i)
		})
	}

	if data != nil {
		x(data)
	}
}

func toMap(tab *lua.LTable, x func(i interface{})) {
	kv := map[string]any{}
	tab.Range(func(key string, val lua.LValue) {
		convert(val, func(i interface{}) {
			kv[key] = i
		})
	})

	x(kv)
}

func convert(v lua.LValue, x func(i interface{})) {
	switch v.Type() {
	case lua.LTInt:
		x(int(v.(lua.LInt)))

	case lua.LTString:
		x(string(v.(lua.LString)))

	case lua.LTBool:
		x(bool(v.(lua.LBool)))

	case lua.LTNumber:
		x(float64(v.(lua.LNumber)))

	case lua.LTTable:
		tab := v.(*lua.LTable)
		if arr := tab.Array(); len(arr) > 0 {
			toArr(arr, x)
			return
		}

		toMap(tab, x)
	}

}

func newConfig(L *lua.LState, tab *lua.LTable) *config {
	cfg := &config{
		Data: make(map[string]any, 6),
	}

	tab.Range(func(key string, val lua.LValue) {
		switch key {
		case "name":
			cfg.Name = auxlib.CheckProcName(val, L)
			return
		case "type":
			cfg.Type = auxlib.CheckProcName(val, L)
			cfg.Data["type"] = auxlib.CheckProcName(val, L)
			return
		}

		convert(val, func(i interface{}) {
			cfg.Data[key] = i
		})
	})

	return cfg
}
