package safecall_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/vela-security/vela-minion/internal/safecall"
)

func doing() error {
	fmt.Println("=====[ doing ]=====")
	return nil
}

func ok() {
	fmt.Println("==[ OK ]==")
}

func onError(err error) {
	fmt.Printf("Error is: %v", err)
}

func TestCall(t *testing.T) {
	safecall.New().Timeout(time.Second).OnComplete(ok).OnError(onError).Exec(doing)
}
