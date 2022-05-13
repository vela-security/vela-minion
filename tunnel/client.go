package tunnel

import (
	"context"
	"errors"
	"github.com/vela-security/vela-public/assert"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/vela-security/vela-minion/internal/banner"

	"github.com/gorilla/websocket"
)

const (
	headerAuthorization   = "Authorization"
	headerWWWAuthenticate = "WWW-Authenticate"
)

type Hide struct {
	Servername string    `json:"servername"`  // wss/https TLS 证书校验时的 servername
	LAN        []string  `json:"lan"`         // broker/manager 的内网地址
	VIP        []string  `json:"vip"`         // broker/manager 的外网地址
	Edition    string    `json:"edition"`     // semver 版本号
	Hash       string    `json:"hash"`        // 文件原始hash
	Size       int       `json:"size"`        // 文件原始size
	DownloadAt time.Time `json:"download_at"` // 下载时间
}

// Client 客户端
type Client struct {
	hide      Hide               // 配置数据
	handler   Handler            // 消息处理器
	xEnv      assert.Environment // 日志打印
	interval  time.Duration      // 心跳间隔
	dialer    *websocket.Dialer  // websocket dialer
	address   address            // 当前所选的连接地址
	status    bool               // 连接状态
	addresses []address          // parse 后的 broker 地址
	conn      *Conn              // 底层 websocket 连接
	client    *http.Client       // http client
	ctx       context.Context    // context
	cancel    context.CancelFunc // cancel
}

func init() {
	banner.Print()
}

// New 新建 client
func New(hide Hide, options ...Option) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := &config{
		interval: 10 * time.Minute,
		dialer:   websocket.DefaultDialer,
		client:   http.DefaultClient,
		ctx:      ctx,
		cancel:   cancel,
	}

	for _, opt := range options {
		opt.apply(cfg)
	}

	return &Client{
		hide:     hide,
		handler:  cfg.handler,
		xEnv:     cfg.env,
		interval: cfg.interval,
		dialer:   cfg.dialer,
		client:   cfg.client,
		ctx:      cfg.ctx,
		cancel:   cfg.cancel,
	}
}

func (c Client) Name() string {
	return "vela.minion.client"
}

func (c Client) Close() error {
	if c.conn != nil {
		_ = c.conn.close()
	}
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

func (c Client) Version() string {
	if conn := c.conn; conn != nil {
		return conn.ident.Edition
	}
	return ""
}

// Start 启动 Client
func (c *Client) Start() error {
	if err := c.pretreatment(); err != nil {
		return err
	}

	err := c.loopBrokerDial()
	if e, ok := err.(*HTTPError); ok && e.Permanently() {
		return e
	}

	// 异步处理消息
	go c.process(err == nil)

	return nil
}

// Inactive client 连接状态
func (c *Client) Inactive() bool {
	return c.conn == nil || !c.status
}

// Push 通过 websocket 通道发送数据
func (c *Client) Push(op assert.Opcode, data interface{}) error {
	if conn := c.conn; conn != nil {
		return conn.Send(&Message{Opcode: op, Data: data})
	}
	return io.ErrUnexpectedEOF
}

// pretreatment 参数预检查与预处理
func (c *Client) pretreatment() error {
	if c.handler == nil {
		return errors.New("handler不能为空")
	}
	if c.hide.Edition == "" {
		return errors.New("edition不能为空")
	}
	lsz, vsz := len(c.hide.LAN), len(c.hide.VIP)
	if lsz == 0 && vsz == 0 {
		return errors.New("lan与vip至少有一条地址")
	}

	servername := c.hide.Servername
	c.addresses = make([]address, 0, lsz+vsz)
	for _, s := range c.hide.LAN {
		var addr address
		if err := addr.parse(false, s, servername); err != nil {
			return err
		}
		c.addresses = append(c.addresses, addr)
	}
	for _, s := range c.hide.VIP {
		var addr address
		if err := addr.parse(true, s, servername); err != nil {
			return err
		}
		c.addresses = append(c.addresses, addr)
	}
	return nil
}

// loopBrokerDial 循环遍历所有连接直至成功
func (c *Client) loopBrokerDial() error {
	var err error
	var conn *Conn
	for _, addr := range c.addresses {
		conn, err = c.dialBroker(addr)
		if err == nil {
			c.conn, c.address, c.status = conn, addr, true
			return nil
		}
		if e, ok := err.(*HTTPError); ok && e.Permanently() {
			return e
		}
		c.xEnv.Warnf("连接 %s 失败: %v", addr.URL, err)
	}
	return err
}

// dialBroker 连接 broker
func (c Client) dialBroker(addr address) (*Conn, error) {
	ident := generateIdent(c.hide.Edition, addr.URL)

	c.xEnv.WithBroker(ident.Arch, ident.MAC, ident.Inet, ident.Inet6, ident.Edition)

	auth, err := ident.marshal()
	if err != nil {
		return nil, err
	}
	header := http.Header{headerAuthorization: []string{auth}}
	minionURL := addr.minionURL()

	c.dialer.TLSClientConfig = addr.TLS
	if tran, ok := c.client.Transport.(*http.Transport); ok {
		tran.TLSClientConfig = addr.TLS
	}

	ws, res, err := c.dial(minionURL, header)
	if err != nil {
		return nil, err
	}

	var claim Claim
	if err = claim.unmarshal(res.Header.Get(headerAuthorization)); err != nil {
		_ = ws.Close()
		return nil, err
	}
	conn := &Conn{conn: ws, ident: ident, claim: claim}

	return conn, nil
}

// dial 连接 websocket
func (c Client) dial(u *url.URL, header http.Header) (*websocket.Conn, *http.Response, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 20*time.Second)
	defer cancel()
	conn, res, err := c.dialer.DialContext(ctx, u.String(), header)
	if err != nil {
		if res == nil {
			return nil, nil, err
		}
		defer func() { _ = res.Body.Close() }()
		buf := make([]byte, 1024)
		n, _ := io.ReadFull(res.Body, buf)
		return nil, nil, &HTTPError{Code: res.StatusCode, Text: string(buf[:n])}
	}

	return conn, res, nil
}

// heartbeat 发送心跳包
func (c Client) heartbeat(ctx context.Context, interval time.Duration) {
	msg := &Message{Opcode: assert.OpHeartbeat}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.xEnv.Debugf("停止发送心跳")
			return
		case <-ticker.C:
			if err := c.conn.Send(msg); err != nil {
				c.xEnv.Warnf("心跳发送失败：%v", err)
			}
		}
	}
}

func (c *Client) process(connected bool) {
	for {
		if connected {
			ident := c.conn.Ident()
			claim := c.conn.Claim()
			c.xEnv.Debugf("连接成功: %s\n节点信息:\nIPv4: %s\nIPv6: %s\nMAC : %s\n版本: %s\n节点ID: %s\nToken: %s\nMask: %d",
				c.address.URL, ident.Inet, ident.Inet6, ident.MAC, ident.Edition, claim.MinionID, claim.Token, claim.Mask)

			// 定时发送心跳
			ctx, cancel := context.WithCancel(c.ctx)
			go c.heartbeat(ctx, c.interval)

			c.handle() // 读取并处理消息，方法会阻塞
			c.status = false
			cancel()          // 取消心跳方法
			connected = false // 状态置为离线
		}

		// 每次重连前要 sleep 5s，留给代理节点足够的时间修改下线状态。
		// 如果客户端断开立即重连，中心端/代理节点对该节点的下线状态未修改完毕，会误认为节点重复登录。
		time.Sleep(5 * time.Second)
		var num int
		for !connected {
			num++
			err := c.loopBrokerDial()
			if connected = err == nil; connected {
				c.xEnv.Infof("第%d次重连成功", num)
				break
			}

			wait := c.wait(num)
			c.xEnv.Infof("%s后第%d次重试", wait, num)
			time.Sleep(wait)
		}
	}
}

// handle 处理消息
func (c *Client) handle() {
	defer func() {
		_ = c.conn.close()
		c.handler.OnDisconnect(c)
	}()
	c.handler.OnConnect(c)

	for {
		rec, err := c.conn.receive()
		if err != nil {
			if isNetClose(err) {
				break
			}
			continue
		}
		c.handler.OnMessage(c, rec)
	}
}

// wait 根据失败次数计算休眠重试事件
func (c Client) wait(num int) time.Duration {
	if num < 50 {
		return 3 * time.Second
	} else if num < 100 {
		return 20 * time.Second
	}
	return time.Minute
}
