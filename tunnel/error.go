package tunnel

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

// HTTPError http 错误
type HTTPError struct {
	Code int    `json:"code"` // http status code
	Text string `json:"text"` // http response body
}

// Error 实现 error 接口
//  401 Unauthorized       : 认证错误
// 		代表客户端发送的认证数据校验有误，一般需要开发者具体分析
// 	403 Forbidden          : 禁止登录
// 		代表客户端已经被标记为删除，不允许登录
//  405 Method Not Allowed : 没有实现
// 		代表服务端出现了错误
//  406 Not Acceptable     : 其它错误
// 		代表其它错误，一般为服务端逻辑问题
//  425 Too Early          : 未激活
// 		代表节点未激活，等待管理员激活即可连接成功
//  429 Too Many Requests  : 重复连接
// 		代表节点已经上线，不允许上线
func (e HTTPError) Error() string {
	switch e.Code {
	case http.StatusUnauthorized:
		return "认证信息错误：" + e.Text
	case http.StatusForbidden:
		return "节点已被禁止登录"
	case http.StatusMethodNotAllowed:
		return "服务器未实现方法"
	case http.StatusNotAcceptable:
		return "服务端认证错误：" + e.Text
	case http.StatusTooEarly:
		return "节点尚未激活"
	case http.StatusTooManyRequests:
		return "重复登录"
	}

	return fmt.Sprintf("status code: %d, error message: %s", e.Code, e.Text)
}

// Permanently 判断错误是否是不可恢复性的错误
//  401 Unauthorized       : 认证错误
// 	403 Forbidden          : 禁止登录
//  405 Method Not Allowed : 没有实现
//  406 Not Acceptable     : 其它错误
//  429 Too Many Requests  : 重复连接
func (e *HTTPError) Permanently() bool {
	return e.Code == http.StatusForbidden
}

func isNetClose(err error) bool {
	switch e := err.(type) {
	case *net.OpError, *websocket.CloseError:
		return true
	case net.Error:
		return e.Timeout()
	default:
		return false
	}
}
