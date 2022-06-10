package minion

import (
	"fmt"
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/buffer"
	"github.com/vela-security/vela-public/kind"
	"github.com/vela-security/vela-public/lua"
)

type ksk struct {
	lua.ProcEx

	topic  string
	object *stream
}

func newKsk(L *lua.LState, object *stream) *ksk {
	topic := L.CheckString(1)
	if e := auxlib.Name(topic); e != nil {
		L.RaiseError("%s invalid topic", object.Name())
		return nil
	}

	return &ksk{
		topic:  topic,
		object: object,
	}
}

func (k *ksk) Name() string {
	return fmt.Sprintf("%s.sdk.%s", k.object.tx.Type(), k.topic)
}

func (k *ksk) Type() string {
	return fmt.Sprintf("%s.sdk", k.object.tx.Type())
}

func (k *ksk) Write(data []byte) (wn int, er error) {
	if !k.object.IsRun() || k.object.socket == nil {
		return
	}

	enc := kind.NewJsonEncoder()
	enc.Tab("")
	enc.KV("topic", k.topic)
	enc.Raw("data", data)
	enc.End("}")
	defer func() {
		buffer.Put(enc.Buffer())
	}()

	wn, er = k.object.socket.Write(enc.Bytes())
	return
}

func (k *ksk) Start() error {
	k.V(lua.PTRun)
	return nil
}

func (k *ksk) Close() error {
	return nil
}
