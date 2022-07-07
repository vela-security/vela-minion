package minion

import (
	"encoding/json"
	"github.com/vela-security/pivot/tasktree"
	opcode "github.com/vela-security/vela-opcode"
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/lua"
)

func pushText(L *lua.LState) int {
	if n := checkEx(L); n != 0 {
		return n
	}

	biz := L.IsInt(1)
	chunk := auxlib.Format(L, 1)
	if len(chunk) == 0 {
		return 0
	}

	err := xEnv.TnlSend(opcode.Opcode(biz), auxlib.S2B(chunk))
	if err != nil {
		L.Push(lua.S2L(err.Error()))
		return 1
	}

	return 0
}

func pushJson(L *lua.LState) int {
	if n := checkEx(L); n != 0 {
		return n
	}

	biz := L.IsInt(1)
	lv := L.Get(2)
	switch lv.Type() {
	case lua.LTNil:
		return 0

	default:
		chunk := lv.String()
		if len(chunk) == 0 {
			return 0
		}

		err := xEnv.TnlSend(opcode.Opcode(biz), json.RawMessage(chunk))
		if err != nil {
			L.Push(lua.S2L(err.Error()))
			return 1
		}
		return 0

	}

}

func pushTaskTree(L *lua.LState) int {
	data := tasktree.ToView()
	err := xEnv.TnlSend(opcode.OpTask, data)
	if err != nil {
		L.Push(lua.S2L(err.Error()))
		return 1
	}
	return 0
}

func (pe *pushEx) Index(L *lua.LState, key string) lua.LValue {
	if lv, ok := pe.meta[key]; ok {
		return lv
	}

	return lua.LNil
}
