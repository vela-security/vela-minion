package minion

import (
	"github.com/vela-security/vela-public/buffer"
	"github.com/vela-security/vela-public/lua"
)

type sub struct {
	tab  *lua.LTable
	code string
}

func newSub(L *lua.LState) *sub {
	return &sub{tab: L.CheckTable(1), code: L.CodeVM()}
}

func (s *sub) Type() string {
	return "sub"
}

func (s *sub) Handle(raw []byte) *buffer.Byte {
	return &buffer.Byte{B: raw}
}

func (s *sub) Config(L *lua.LState) *config {
	return newConfig(L, s.tab)
}

func (s *sub) Sdk(L *lua.LState, st *stream) *lua.ProcData {
	return nil
}

func (s *sub) CodeVM() string {
	return s.code
}

func newStreamSub(L *lua.LState) int {
	return help(L, newSub(L))
}
