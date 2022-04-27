package rockcli

import (
	"crypto/tls"
	"errors"
	"net/url"
	"strings"
)

type address struct {
	VIP   bool
	URL   *url.URL
	IsTLS bool
	TLS   *tls.Config
}

func (a address) minionURL() *url.URL {
	u := a.URL
	return &url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path + "/v1/minion/endpoint", RawQuery: u.RawQuery}
}

func (a address) streamURL() *url.URL {
	u := a.URL
	return &url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path + "/v1/minion/stream", RawQuery: u.RawQuery}
}

func (a address) appendURL(path string) *url.URL {
	u := a.URL
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	return &url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path + path, RawQuery: u.RawQuery}
}

func (a address) appendToHTTP(path, query string) *url.URL {
	u := a.URL
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	scheme := "http"
	if a.IsTLS {
		scheme = "https"
	}
	rqu := u.RawQuery
	if rqu != "" {
		rqu += "&"
	}
	rqu += query

	return &url.URL{Scheme: scheme, Host: u.Host, Path: u.Path + path, RawQuery: rqu}
}

func (a *address) parse(vip bool, u, servername string) error {
	pu, err := url.Parse(u)
	if err != nil {
		return err
	}
	scheme := strings.ToLower(pu.Scheme)
	a.URL, a.IsTLS, a.VIP = pu, scheme == "wss", vip
	if a.IsTLS && servername != "" {
		a.TLS = &tls.Config{ServerName: servername}
	}

	if scheme != "ws" && scheme != "wss" {
		return errors.New("必须是ws或wss的url")
	}
	return nil
}
