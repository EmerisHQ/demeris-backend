package rpcwatcher

import (
	"time"
)

type watchdog struct {
	timeout       chan struct{}
	ping          chan struct{}
	timer         *time.Timer
	timeoutAmount time.Duration
}

func newWatchdog(timeoutAmount time.Duration) *watchdog {
	return &watchdog{
		timeout:       make(chan struct{}),
		ping:          make(chan struct{}),
		timeoutAmount: timeoutAmount,
	}
}

func (w watchdog) Ping() {
	go func() {
		w.ping <- struct{}{}
	}()
}

func (w *watchdog) fireTimeoutSignal() {
	w.timeout <- struct{}{}
}

func (w *watchdog) Start() {
	w.timer = time.AfterFunc(w.timeoutAmount, w.fireTimeoutSignal)

	go func() {
		for {
			select {
			case <-w.ping:
				if !w.timer.Stop() {
					<-w.timer.C
				}

				w.timer = time.AfterFunc(w.timeoutAmount, w.fireTimeoutSignal)
			}
		}
	}()
}
