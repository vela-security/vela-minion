package minion

import (
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-public/lua"
)

var xEnv assert.Environment

func WithEnv(env assert.Environment) {
	xEnv = env
	env.Set("push", newPushEx())
	uv := lua.NewUserKV()
	uv.Set("new", lua.NewFunction(newStreamSub))
	uv.Set("kfk", lua.NewFunction(newStreamKfk))
	uv.Set("tcp", lua.NewFunction(newStreamTcp))
	env.Set("stream", uv)
}
