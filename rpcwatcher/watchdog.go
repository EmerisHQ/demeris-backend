package rpcwatcher

import (
	"time"
)

type watchdog struct {
	timeout       chan bool
	ping          chan struct{}
	timer         *time.Timer
	timeoutAmount time.Duration
}

func newWatchdog(timeoutAmount time.Duration) *watchdog {
	return &watchdog{
		timeout:       make(chan bool),
		ping:          make(chan struct{}),
		timeoutAmount: timeoutAmount,
	}
}

func (w watchdog) Ping() {
	go func() {
		w.ping <- struct{}{}
	}()
}

func (w *watchdog) Start() {
	w.timer = time.AfterFunc(w.timeoutAmount, func() {
		w.timeout <- false
	})

	go func() {
		for {
			select {
			case <-w.ping:
				if !w.timer.Stop() {
					<-w.timer.C
				}

				w.timer = time.AfterFunc(w.timeoutAmount, func() {
					w.timeout <- true
				})
			}
		}
	}()
}

func (w watchdog) ReadTimeout(watcher *Watcher) {
	for {
		select {
		case <-w.timeout:
			resubscribe(watcher)
		}
	}
}
