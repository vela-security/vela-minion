package tunnel

import "github.com/vela-security/vela-minion/internal/ciphertext"

type Claim struct {
	MinionID string `json:"minion_id"`
	Mask     byte   `json:"mask"`
	Token    string `json:"token"`
}

func (c Claim) marshal() (string, error) {
	raw, err := ciphertext.EncryptJSON(c)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (c *Claim) unmarshal(enc string) error {
	return ciphertext.DecryptJSON([]byte(enc), c)
}
