package tunnel

import (
	"net"
	"net/url"
	"runtime"
	"strings"

	"github.com/vela-security/vela-minion/internal/ciphertext"
)

type Ident struct {
	Inet    net.IP           `json:"inet"    validate:"required"`                   // 本机 IPv4
	Inet6   net.IP           `json:"inet6"   validate:"required"`                   // 本机 IPv6
	MAC     net.HardwareAddr `json:"mac"     validate:"required"`                   // MAC
	Goos    string           `json:"goos"    validate:"oneof=linux windows darwin"` // runtime.GOOS
	Arch    string           `json:"arch"    validate:"oneof=amd64 386 arm64 arm"`  // runtime.GOARCH
	Edition string           `json:"edition" validate:"semver"`                     // 版本号
}

// generateIdent 构造 minion 节点的身份认证信息
func generateIdent(edition string, dest *url.URL) Ident {
	ident := Ident{Goos: runtime.GOOS, Arch: runtime.GOARCH, Edition: edition}
	ident.complementNet(dest) // 补全网络信息
	return ident
}

func (i Ident) marshal() (string, error) {
	raw, err := ciphertext.EncryptJSON(i)
	return string(raw), err
}

// complementNet 补全网络信息
func (i *Ident) complementNet(u *url.URL) {
	if u == nil || u.Host == "" {
		return
	}
	host := u.Host
	if _, _, err := net.SplitHostPort(host); err != nil {
		if strings.ToLower(u.Scheme) == "wss" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	conn, err := net.Dial("udp", host)
	if err != nil {
		return
	}
	_ = conn.Close()

	// 此时我们并不知道这是 ipv4 还是 ipv6
	ip := conn.LocalAddr().(*net.UDPAddr).IP
	if ip.To4() != nil {
		i.Inet = ip.To4()
	} else {
		i.Inet6 = ip.To16()
	}

	anther, mac := i.anotherInet(ip)
	i.MAC = mac

	if ip.Equal(anther) { //判断anther 是否相等
		return
	}

	if anther.To4() != nil {
		i.Inet = anther.To4()
	} else {
		i.Inet6 = anther.To16()
	}
}

// anotherAddr 查询ip mac地址
func (i Ident) anotherInet(ip net.IP) (net.IP, net.HardwareAddr) {
	faces, _ := net.Interfaces()
	for _, face := range faces {
		addr, _ := face.Addrs()
		if inet, ok := i.lookupIP(ip, addr); ok {
			return inet, face.HardwareAddr
		}
	}
	return nil, nil
}

func (Ident) lookupIP(ip net.IP, addr []net.Addr) (net.IP, bool) {
	v4 := ip.To4() != nil
	for _, n := range addr {
		if n.(*net.IPNet).IP.Equal(ip) {
			for _, ad := range addr {
				inet := ad.(*net.IPNet).IP
				if inet.To4() == nil && !v4 {
					return inet.To16(), true
				}
				if inet.To4() != nil && v4 {
					return inet.To4(), true
				}
			}
		}
	}
	return nil, false
}

// streamIdent stream 模式的认证包
type streamIdent struct {
	Mode string `json:"mode"` // Stream 的连接模式
	Data any    `json:"data"` // 内部数据
}

func (i streamIdent) marshal() (string, error) {
	enc, err := ciphertext.EncryptJSON(i)
	return string(enc), err
}
