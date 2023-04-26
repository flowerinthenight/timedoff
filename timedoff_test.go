package timedoff

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	var n int32
	now := time.Now()
	done := make(chan error, 1)
	to := New(time.Second*3, &CallbackT{
		Callback: func(p interface{}) {
			atomic.StoreInt32(&n, p.(int32))
			done <- nil
		},
		Params: int32(1),
	})

	go func() {
		time.Sleep(time.Second * 2)
		to.On() // reset back to 3s
	}()

	<-done // wait for timer
	since := time.Since(now)

	if since.Seconds() < 5 {
		t.Logf("expected 5+ seconds, got %v", since.Seconds())
		t.Fail()
	}

	if atomic.LoadInt32(&n) != 1 {
		t.Logf("params: expected 1, got %v", atomic.LoadInt32(&n))
		t.Fail()
	}
}
