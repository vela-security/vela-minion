package minion

import (
	"github.com/vela-security/vela-public/buffer"
	"github.com/vela-security/vela-public/lua"
)

type Tx interface {
	CodeVM() string
	Type() string

	Sdk(*lua.LState, *stream) *lua.ProcData
	Handle([]byte) *buffer.Byte
	Config(*lua.LState) *config
}
