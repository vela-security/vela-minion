package minion

import (
	"github.com/vela-security/vela-public/lua"
)

func help(L *lua.LState, tx Tx) int {
	cfg := tx.Config(L)
	proc := L.NewProc(cfg.Name, streamTypeOf)

	if proc.IsNil() {
		proc.Set(newStream(cfg, tx))
	} else {
		proc.Data.(*stream).cfg = cfg
		proc.Data.(*stream).tx = tx
	}
	L.Push(proc)
	return 1
}

func (st *stream) pushL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	pn := 0
	for i := 1; i <= n; i++ {
		wn, err := st.Write(lua.S2B(L.Get(i).String()))
		pn += wn

		if err != nil {
			L.Push(lua.LNumber(pn))
			L.Pushf("%v", err)
			return 2
		}
	}

	L.Push(lua.LInt(pn))
	L.Push(lua.LNil)
	return 2
}

func (st *stream) startL(L *lua.LState) int {
	xEnv.Start(L, st).From(st.tx.CodeVM()).Do()
	return 0
}

func (st *stream) sdkL(L *lua.LState) int {
	L.Push(st.tx.Sdk(L, st))
	return 1
}

func (st *stream) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "sdk":
		return L.NewFunction(st.sdkL)

	case "push":
		return L.NewFunction(st.pushL)

	case "start":
		return L.NewFunction(st.startL)

	default:
		//todo
	}

	return lua.LNil
}
