package minion

import (
	"fmt"
	"github.com/vela-security/vela-public/buffer"
	"github.com/vela-security/vela-public/kind"
	"github.com/vela-security/vela-public/lua"
)

type kfk struct {
	tab   *lua.LTable
	code  string
	topic []byte
}

func newKafka(L *lua.LState) *kfk {
	tab := L.CheckTable(1)
	tab.RawSetString("type", lua.S2L("kafka"))
	return &kfk{tab: tab, code: L.CodeVM()}

}

func (k *kfk) Topic() ([]byte, error) {
	if len(k.topic) != 0 {
		return k.topic, nil
	}

	val := k.tab.RawGetString("topic")
	if val.Type() != lua.LTString {
		return nil, fmt.Errorf("invalid topic type , got %s", val.Type().String())
	}

	topic := lua.S2B(val.String())
	if len(topic) > 255 {
		return nil, fmt.Errorf("topic length > 255")
	}

	k.topic = topic
	return topic, nil
}

func (k *kfk) Type() string {
	return "kafka"
}

/*
	frame {
		topic string
		data  string
	}
*/
func (k *kfk) Handle(raw []byte) *buffer.Byte {
	enc := kind.NewJsonEncoder()
	enc.Tab("")
	enc.KV("topic", k.topic)
	enc.Raw("data", raw)
	enc.End("}")

	return enc.Buffer()
}

func (k *kfk) Config(L *lua.LState) *config {

	if _, err := k.Topic(); err != nil {
		L.RaiseError("invalid stream kafka topic")
		return nil
	}

	return newConfig(L, k.tab)
}

func (k *kfk) CodeVM() string {
	return k.code
}

func (k *kfk) Sdk(L *lua.LState, st *stream) *lua.ProcData {
	return L.NewProcData(newKsk(L, st))
}

func newStreamKfk(L *lua.LState) int {
	return help(L, newKafka(L))
}
