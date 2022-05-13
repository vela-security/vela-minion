package minion

import (
	"github.com/vela-security/pivot/tasktree"
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/lua"
)

func pushL(L *lua.LState) int {
	if n := checkEx(L); n != 0 {
		return n
	}

	biz := L.IsInt(1)
	chunk := auxlib.Format(L, 1)
	if len(chunk) == 0 {
		return 0
	}

	err := xEnv.TnlSend(assert.Opcode(biz), auxlib.S2B(chunk))
	if err != nil {
		L.Push(lua.S2L(err.Error()))
		return 1
	}

	return 0
}

func pushTaskTree(L *lua.LState) int {
	data := tasktree.ToView()
	err := xEnv.TnlSend(assert.OpCoreService, data)
	if err != nil {
		L.Push(lua.S2L(err.Error()))
		return 1
	}
	return 0
}

func (pe *pushEx) Index(L *lua.LState, key string) lua.LValue {
	if key == "any" {
		return pe.any
	}

	if lv, ok := pe.meta[key]; ok {
		return lv
	}

	return lua.LNil
}
