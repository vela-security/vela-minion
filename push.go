package minion

import (
	"encoding/json"
	audit "github.com/vela-security/vela-audit"
	opcode "github.com/vela-security/vela-opcode"
	"github.com/vela-security/vela-public/lua"
)

type pushEx struct {
	any  lua.LValue
	meta map[string]lua.LValue
}

func (pe *pushEx) String() string                         { return "rock.push.exdata" }
func (pe *pushEx) Type() lua.LValueType                   { return lua.LTObject }
func (pe *pushEx) AssertFloat64() (float64, bool)         { return 0, false }
func (pe *pushEx) AssertString() (string, bool)           { return "", false }
func (pe *pushEx) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (pe *pushEx) Peek() lua.LValue                       { return pe }

func newPushEx() *pushEx {
	return &pushEx{
		meta: map[string]lua.LValue{
			"sysinfo":    newPushFunc(opcode.OpSysInfo),
			"cpu":        newPushFunc(opcode.OpCPU),
			"disk":       newPushFunc(opcode.OpDiskIO),
			"memory":     newPushFunc(opcode.OpMemory),
			"socket":     newPushFunc(opcode.OpSocket),
			"network":    newPushFunc(opcode.OpNetwork),
			"service":    newPushFunc(opcode.OpService),
			"filesystem": newPushFunc(opcode.OpFileSystem),
			"task":       lua.NewFunction(pushTaskTree),
			"json":       lua.NewFunction(pushJson),
			"text":       lua.NewFunction(pushText),
		}}
}

func checkEx(L *lua.LState) int {
	if xEnv.TnlIsDown() {
		L.Push(lua.S2L("tunnel client is down"))
		return 1
	}
	return 0
}

func pushExec(op opcode.Opcode, L *lua.LState) int {

	if n := checkEx(L); n > 0 {
		return n
	}

	var err error
	val := L.Get(1)

	switch val.Type() {

	case lua.LTString, lua.LTObject, lua.LTMap, lua.LTSlice:
		raw := val.String()
		err = xEnv.TnlSend(op, json.RawMessage(raw))
		if err != nil {
			audit.Errorf("push %s fail %v", op.String(), err).From(L.CodeVM()).Log().Put()
			return 0
		}

	default:
		L.Pushf("invalid type %s", val.Type().String())
		return 1
	}

	return 0
}

func newPushFunc(op opcode.Opcode) *lua.LFunction {
	return lua.NewFunction(func(L *lua.LState) int {
		return pushExec(op, L)
	})
}
