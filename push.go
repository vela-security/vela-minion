package minion

import (
	"encoding/json"
	audit "github.com/vela-security/vela-audit"
	"github.com/vela-security/vela-public/assert"
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
		any: lua.NewFunction(pushL),
		meta: map[string]lua.LValue{
			"sysinfo":    newPushFunc(assert.OpSysInfo),
			"cpu":        newPushFunc(assert.OpCPU),
			"disk":       newPushFunc(assert.OpDiskIO),
			"listen":     newPushFunc(assert.OpListen),
			"memory":     newPushFunc(assert.OpMemory),
			"socket":     newPushFunc(assert.OpSocket),
			"network":    newPushFunc(assert.OpNetwork),
			"process":    newPushFunc(assert.OpProcess),
			"service":    newPushFunc(assert.OpService),
			"account":    newPushFunc(assert.OpAccount),
			"filesystem": newPushFunc(assert.OpFileSystem),
			"task":       lua.NewFunction(pushTaskTree),
		}}
}

func checkEx(L *lua.LState) int {
	if xEnv.TnlIsDown() {
		L.Push(lua.S2L("tunnel client is down"))
		return 1
	}
	return 0
}

func pushExec(op assert.Opcode, L *lua.LState) int {
	if n := checkEx(L); n > 0 {
		return n
	}

	var err error
	val := L.Get(1)

	switch val.Type() {

	case lua.LTString, lua.LTObject:
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

func newPushFunc(op assert.Opcode) *lua.LFunction {
	return lua.NewFunction(func(L *lua.LState) int {
		return pushExec(op, L)
	})
}
