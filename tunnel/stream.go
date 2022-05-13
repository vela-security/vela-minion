package tunnel

import (
	"io"
	"net/http"
)

// Stream 建立 stream 通道
func (c Client) Stream(mode string, data any) (*StreamConn, error) {
	conn := c.conn
	if conn == nil {
		return nil, io.EOF
	}

	ident := &streamIdent{Mode: mode, Data: data}
	auth, err := ident.marshal()
	if err != nil {
		return nil, err
	}

	token := conn.Claim().Token
	header := http.Header{headerAuthorization: []string{token}, headerWWWAuthenticate: []string{auth}}

	streamURL := c.address.streamURL()
	ws, _, err := c.dial(streamURL, header)
	if err != nil {
		return nil, err
	}
	stream := newStreamConn(ws)

	return stream, nil
}
