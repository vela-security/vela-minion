package minion

import (
	"github.com/vela-security/vela-public/buffer"
	"github.com/vela-security/vela-public/lua"
)

type tcp struct {
	tab  *lua.LTable
	code string
}

func newTcp(L *lua.LState) *tcp {
	tab := L.CheckTable(1)
	tab.RawSetString("type", lua.S2L("forward"))
	tab.RawSetString("network", lua.S2L("tcp"))
	return &tcp{tab: tab, code: L.CodeVM()}
}

func (t *tcp) Type() string {
	return "tcp.forward"
}
func (t *tcp) Handle(raw []byte) *buffer.Byte {
	return &buffer.Byte{B: raw}
}

func (t *tcp) Config(L *lua.LState) *config {
	return newConfig(L, t.tab)
}

func (t *tcp) CodeVM() string {
	return t.code
}

func (t *tcp) Sdk(L *lua.LState, s *stream) *lua.ProcData {
	return nil
}

func newStreamTcp(L *lua.LState) int {
	return help(L, newTcp(L))
}
