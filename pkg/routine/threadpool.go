// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package routine

import (
	"context"
	"errors"
	"sync"
)

// ThreadPool implements a thread pool with limited pool size
type ThreadPool struct {
	tasks     chan Task
	mutex     sync.Mutex
	limiter   chan interface{}
	wg        sync.WaitGroup
	terminate chan interface{}
}

// NewThreadPool creates a new thread pool
func NewThreadPool(size int) (*ThreadPool, error) {
	if size <= 0 {
		return nil, errors.New("invalid thread pool size")
	}
	return &ThreadPool{
		tasks:     make(chan Task),
		limiter:   make(chan interface{}, size),
		terminate: make(chan interface{}, 1),
	}, nil
}

// Start starts the thread pool
func (tp *ThreadPool) Start(context.Context) error {
	ready := make(chan struct{})
	go func() {
		close(ready)
		select {
		case <-tp.terminate:
			return
		case t := <-tp.tasks:
			tp.execute(t)
		}
	}()
	<-ready

	return nil
}

// Stop stops the thread pool, and wait until all tasks in the pool finished
func (tp *ThreadPool) Stop(context.Context) error {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()
	tp.terminate <- struct{}{}
	close(tp.tasks)
	for t := range tp.tasks {
		tp.execute(t)
	}
	tp.wg.Wait()

	return nil
}

// Add adds a task into the pool
func (tp *ThreadPool) Add(t Task) {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()
	tp.tasks <- t
}

func (tp *ThreadPool) execute(t Task) {
	tp.wg.Add(1)
	tp.limiter <- true
	go func() {
		defer func() {
			tp.wg.Done()
			<-tp.limiter
		}()
		t()
	}()
}
