package timedoff

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type CallbackT struct {
	Callback func(interface{})
	Params   interface{}
}

// TimedOff is an object that we can use as a generic atomic on/off switch that
// automatically turns itself off after a specified time. Similar to a time.Timer
// object but built for concurrent use.
type TimedOff struct {
	mtx      *sync.Mutex
	on       int32
	duration int64 // in seconds
	cb       *CallbackT
	ctx      context.Context
	cancel   context.CancelFunc
	ch       chan error
}

func (t *TimedOff) IsOn() bool { return atomic.LoadInt32(&t.on) == 1 }
func (t *TimedOff) Off()       { atomic.StoreInt32(&t.on, 0) }

// On resets the internal timer.
func (t *TimedOff) On() {
	if atomic.LoadInt32(&t.on) == 0 {
		go t.run()
	}

	t.ch <- nil // reset
}

func (t *TimedOff) setDeadline() {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.ctx, t.cancel = context.WithTimeout(
		context.Background(),
		time.Second*time.Duration(t.duration),
	)
}

func (t *TimedOff) run() {
	t.setDeadline()
	atomic.StoreInt32(&t.on, 1)

loop:
	for {
		select {
		case <-t.ch:
			t.cancel()
			t.setDeadline()
		case <-t.ctx.Done():
			if t.cb != nil {
				t.cb.Callback(t.cb.Params)
			}

			atomic.StoreInt32(&t.on, 0)
			break loop
		}
	}
}

// New creates a TimedOff object with 'duration' in seconds.
func New(duration int64, cb ...*CallbackT) *TimedOff {
	secs := duration
	if secs == 0 {
		secs = 30
	}

	to := TimedOff{
		mtx:      &sync.Mutex{},
		duration: secs,
		ch:       make(chan error),
	}

	if len(cb) > 0 {
		to.cb = cb[0]
	}

	go to.run() // start
	return &to
}
