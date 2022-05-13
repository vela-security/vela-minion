package tunnel

import (
	"github.com/vela-security/vela-public/assert"
)

type Handler interface {
	OnConnect(*Client)
	OnMessage(*Client, *Receive)
	OnDisconnect(*Client)
}

type noopHandler struct {
	xEnv assert.Environment
}

func (n noopHandler) OnConnect(cli *Client) {
	n.xEnv.Infof("节点连接成功")
}

func (n noopHandler) OnMessage(cli *Client, rec *Receive) {
	n.xEnv.Infof("收到了消息：%s", rec.Opcode())
}

func (n noopHandler) OnDisconnect(cli *Client) {
	n.xEnv.Infof("节点断开了连接")
}
