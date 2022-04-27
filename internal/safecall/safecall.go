package safecall

import (
	"context"
	"time"
)

type builder struct {
	timeout    time.Duration
	execFn     func() error
	completeFn func()
	errorFn    func(error)
	panicFn    func(any)
	timeoutFn  func()
}

func New() *builder {
	return &builder{}
}

func (b *builder) Timeout(timeout time.Duration) *builder {
	b.timeout = timeout
	return b
}

func (b *builder) OnComplete(fn func()) *builder {
	b.completeFn = fn
	return b
}

func (b *builder) OnError(fn func(error)) *builder {
	b.errorFn = fn
	return b
}

func (b *builder) OnPanic(fn func(any)) *builder {
	return b
}

func (b *builder) OnTimeout(fn func()) *builder {
	b.timeoutFn = fn
	return b
}

func (b *builder) Exec(fn func() error) {
	if fn == nil {
		return
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if b.timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), b.timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	var err error
	var panicked bool
	go func() {
		defer func() {
			cancel()
			if cause := recover(); cause != nil {
				panicked = true
				if b.panicFn != nil {
					b.panicFn(cause)
				}
			}
		}()
		if err = fn(); err != nil && b.errorFn != nil {
			b.errorFn(err)
		}
	}()

	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		if b.timeoutFn != nil {
			b.timeoutFn()
		}
	} else {
		if err == nil && !panicked && b.completeFn != nil {
			b.completeFn()
		}
	}
}
