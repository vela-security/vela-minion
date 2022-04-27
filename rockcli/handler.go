package rockcli

import "github.com/vela-security/vela-minion/internal/logger"

type Handler interface {
	OnConnect(*Client)
	OnMessage(*Client, *Receive)
	OnDisconnect(*Client)
}

type noopHandler struct {
	logger logger.Logger
}

func (n noopHandler) OnConnect(cli *Client) {
	n.logger.Infof("节点连接成功")
}

func (n noopHandler) OnMessage(cli *Client, rec *Receive) {
	n.logger.Infof("收到了消息：%s", rec.Opcode())
}

func (n noopHandler) OnDisconnect(cli *Client) {
	n.logger.Infof("节点断开了连接")
}
