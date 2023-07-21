/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package interrupts exposes helpers for graceful handling of interrupt signals
package interrupts

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/beego/beego/v2/core/logs"
)

// only one instance of the manager ever exists
var single *manager

func init() {
	m := sync.Mutex{}
	single = &manager{
		c:  sync.NewCond(&m),
		wg: sync.WaitGroup{},
	}
	go handleInterrupt()
}

type manager struct {
	// only one signal handler should be installed, so we use a cond to
	// broadcast to workers that an interrupt has occurred
	c *sync.Cond
	// we record whether we've broadcast in the past
	seenSignal bool
	// we want to ensure that all registered servers and workers get a
	// change to gracefully shut down
	wg sync.WaitGroup
}

// handleInterrupt turns an interrupt into a broadcast for our condition.
// This must be called _first_ before any work is registered with the
// manager, or there will be a deadlock.
func handleInterrupt() {
	signalsLock.Lock()
	sigChan := signals()
	signalsLock.Unlock()
	s := <-sigChan
	logs.Info("Received signal: %v.", s)
	single.c.L.Lock()
	single.seenSignal = true
	single.c.Broadcast()
	single.c.L.Unlock()
}

// test initialization will set the signals channel in another goroutine
// so we need to synchronize that in order to not trigger the race detector
// even though we know that init() calls will be serial and the test init()
// will fire first
var signalsLock = sync.Mutex{}

// signals allows for injection of mock signals in testing
var signals = func() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	return sig
}

// wait executes the cancel when an interrupt is seen or if one has already
// been handled
func wait(cancel func()) {
	single.c.L.Lock()
	if !single.seenSignal {
		single.c.Wait()
	}
	single.c.L.Unlock()
	cancel()
}

// WaitForGracefulShutdown waits until all registered servers and workers
// have had time to gracefully shut down, or times out. This function is
// blocking.
func WaitForGracefulShutdown() {
	wait(func() {
		logs.Info("Interrupt received.")
	})

	single.wg.Wait()

	logs.Info("All workers gracefully terminated, exiting.")
}

// OnInterrupt ensures that work is done when an interrupt is fired
// and that we wait for the work to be finished before we consider
// the process cleaned up. This function is not blocking.
func OnInterrupt(work func()) {
	single.wg.Add(1)
	go wait(func() {
		work()
		single.wg.Done()
	})
}
