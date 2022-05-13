package tunnel

import (
	"context"
	"github.com/vela-security/vela-public/assert"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type config struct {
	handler  Handler
	interval time.Duration
	env      assert.Environment
	dialer   *websocket.Dialer
	client   *http.Client
	ctx      context.Context
	cancel   context.CancelFunc
}

// An Option configures a Client.
type Option interface {
	apply(*config)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*config)

func (f optionFunc) apply(cfg *config) {
	f(cfg)
}

func WithInterval(d time.Duration) Option {
	return optionFunc(func(cfg *config) {
		if d > time.Second && d <= time.Hour {
			cfg.interval = d
		}
	})
}

func WithEnv(env assert.Environment) Option {
	return optionFunc(func(cfg *config) {
		if env != nil {
			cfg.env = env
			cfg.handler = noopHandler{env}
		}
	})
}

func WithDialer(dialer *websocket.Dialer) Option {
	return optionFunc(func(cfg *config) {
		if dialer != nil {
			cfg.dialer = dialer
		}
	})
}

func WithHandler(handler Handler) Option {
	return optionFunc(func(cfg *config) {
		if handler != nil {
			cfg.handler = handler
		}
	})
}

func WithContext(ctx context.Context) Option {
	return optionFunc(func(cfg *config) {
		if ctx != nil {
			if cfg.cancel != nil {
				cfg.cancel()
			}
			cfg.ctx, cfg.cancel = context.WithCancel(ctx)
		}
	})
}

func WithHTTPClient(client *http.Client) Option {
	return optionFunc(func(cfg *config) {
		if client != nil {
			cfg.client = client
		}
	})
}
