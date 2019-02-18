package queue

import (
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// Queue handles document indexing
type Queue struct {
	l *zap.SugaredLogger

	fmux      sync.RWMutex
	closeFunc func()
	flushFunc func()

	pendingC  chan func()
	pending   *int32
	rate      time.Duration
	batchSize int32

	stopC   chan bool
	stopped bool
}

// Options declares queue mangement options
type Options struct {
	Rate      time.Duration
	BatchSize int32
}

// New instantiates a new queue. flushFunc is is used for periodic index flusing,
// and closeFunc will be used when closing. closeFunc should also handle a flush.
func New(
	logger *zap.SugaredLogger,
	flushFunc, closeFunc func(),
	opts Options,
) *Queue {
	if flushFunc == nil {
		flushFunc = func() {}
	}
	if closeFunc == nil {
		closeFunc = func() {}
	}
	if opts.Rate == 0 {
		opts.Rate = 5 * time.Second
	}

	var counter = int32(0)
	return &Queue{
		l: logger,

		closeFunc: closeFunc,
		flushFunc: flushFunc,

		pendingC:  make(chan func(), 3*opts.BatchSize),
		pending:   &counter,
		rate:      opts.Rate,
		batchSize: opts.BatchSize,

		stopC:   make(chan bool, 1),
		stopped: false,
	}
}

// Queue indicates that a new item is pending insertion
func (q *Queue) Queue(action func()) {
	if !q.stopped {
		q.pendingC <- action
	} else {
		q.l.Error("queue failed: queue is stopped, cannot queue more elements")
	}
}

// Run maintains the queue and executes flushes as necessary
func (q *Queue) Run() {
	var ticker = time.NewTicker(q.rate)
	for {
		select {
		case <-ticker.C:
			q.l.Debugw("preparing to flush")
			q.flush()

		case action := <-q.pendingC:
			q.l.Debugw("executing and adding to flush queue")
			go q.queue(action)

		case <-q.stopC:
			q.l.Infow("stopping background job")
			ticker.Stop()
			return
		}
	}
}

// RLock allows engine to make queries without interfering with indexing ops
// TODO: investigate Riot search concurrency, currently bugged (https://github.com/go-ego/riot/issues/82)
func (q *Queue) RLock() { q.fmux.Lock() }

// RUnlock allows engine to make queries without interfering with indexing ops
// TODO: investigate Riot search concurrency, currently bugged (https://github.com/go-ego/riot/issues/82)
func (q *Queue) RUnlock() { q.fmux.Unlock() }

// Close stops the queue runner and releases queue assets
func (q *Queue) Close() {
	q.stopC <- true
	q.stop()
}

func (q *Queue) queue(action func()) {
	q.fmux.Lock()
	action()
	atomic.AddInt32(q.pending, 1)
	q.fmux.Unlock()
}

func (q *Queue) flush() {
	if pending := atomic.LoadInt32(q.pending); pending > q.batchSize {
		q.fmux.Lock()

		q.l.Infow("executing flush", "items", pending)
		var now = time.Now()
		q.flushFunc()
		atomic.StoreInt32(q.pending, 0)
		q.l.Infow("flush complete",
			"items", pending,
			"duration", time.Since(now))

		q.fmux.Unlock()
	}
}

func (q *Queue) stop() {
	q.fmux.Lock()

	var pending = atomic.LoadInt32(q.pending)
	q.l.Infow("executing close",
		"items", pending)
	var now = time.Now()
	q.closeFunc()
	q.l.Infow("flush complete",
		"items", pending,
		"duration", time.Since(now))

	// prevent further entries
	q.stopped = true

	q.fmux.Unlock()
}
