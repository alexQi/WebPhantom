package scheduler

import (
	"fmt"
	"golang.org/x/time/rate"
	"sync"
	"sync/atomic"
	"time"
)

type worker struct {
	id          string
	queue       string
	taskChan    chan *Task
	quitChan    chan struct{}
	limiter     *rate.Limiter
	active      int32
	lastActive  time.Time
	idleTimeout time.Duration
	once        sync.Once // 新增：保护关闭
}

func newWorker(queue string, idleTimeout time.Duration, qps int) *worker {
	return &worker{
		id:          fmt.Sprintf("%s-%d", queue, time.Now().UnixNano()),
		queue:       queue,
		taskChan:    make(chan *Task, 100),
		quitChan:    make(chan struct{}),
		limiter:     rate.NewLimiter(rate.Every(time.Minute/time.Duration(qps)), 1),
		lastActive:  time.Now(),
		idleTimeout: idleTimeout,
	}
}

func (w *worker) IsActive() bool {
	return atomic.LoadInt32(&w.active) == 1
}

func (w *worker) stop() {
	w.once.Do(func() {
		close(w.quitChan)
	})
}
