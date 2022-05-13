package tunnel

import (
	"encoding/binary"
	"encoding/json"
	"github.com/vela-security/vela-public/assert"
)

type Message struct {
	Opcode assert.Opcode
	Data   any
}

func (m Message) marshal(mask byte) ([]byte, error) {
	return marshal(uint16(m.Opcode), m.Data, mask)
}

type Receive struct {
	opcode assert.Opcode
	data   []byte
}

// Opcode 操作码
func (r Receive) Opcode() assert.Opcode {
	return r.opcode
}

// Bind 参数绑定
func (r Receive) Bind(v any) error {
	if len(r.data) == 0 {
		return nil
	}
	return json.Unmarshal(r.data, v)
}

func (r *Receive) unmarshal(raw []byte, mask byte) error {
	opcode, data, err := unmarshal(raw, mask)
	if err != nil {
		return err
	}
	r.opcode, r.data = assert.Opcode(opcode), data
	return nil
}

func unmarshal(raw []byte, mask byte) (uint16, []byte, error) {
	if len(raw) < 2 {
		return 0, nil, &json.SyntaxError{Offset: 2}
	}
	for i := range raw {
		raw[i] ^= mask
	}
	opcode := binary.BigEndian.Uint16(raw)

	return opcode, raw[2:], nil
}

func marshal(opcode uint16, data any, mask byte) ([]byte, error) {
	var dat []byte
	if data != nil {
		var err error
		if dat, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}

	raw := make([]byte, 2+len(dat))
	binary.BigEndian.PutUint16(raw, opcode)
	if dat != nil {
		copy(raw[2:], dat)
	}

	for i := range raw {
		raw[i] ^= mask
	}

	return raw, nil
}
