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

	// fmux is RLocked for pending increments, and Locked during flushes, so we
	// don't increment during a flush
	fmux        *sync.RWMutex
	closeFunc   func()
	flushFunc   func()
	forceFlushC chan bool

	pendingC  chan string
	pending   *int32
	rate      time.Duration
	batchSize int32

	stopC   chan bool
	stopped bool
}

// New instantiates a new queue. flushFunc is is used for periodic index flusing,
// and closeFunc will be used when closing. closeFunc should also handle a flush.
func New(
	logger *zap.SugaredLogger,
	flushFunc, closeFunc func(),
	rate time.Duration,
) *Queue {
	if flushFunc == nil {
		flushFunc = func() {}
	}
	if closeFunc == nil {
		closeFunc = func() {}
	}
	if rate == 0 {
		rate = 5 * time.Second
	}

	var counter = int32(0)
	return &Queue{
		l: logger,

		closeFunc:   closeFunc,
		flushFunc:   flushFunc,
		forceFlushC: make(chan bool),

		pendingC:  make(chan string),
		pending:   &counter,
		rate:      rate,
		batchSize: 3,

		stopC:   make(chan bool, 1),
		stopped: false,
	}
}

// Queue indicates that a new item is pending insertion
func (q *Queue) Queue(hash string) {
	if !q.stopped {
		q.pendingC <- hash
	} else {
		q.l.Errorw("queue is stopped, can not queue more elements",
			"hash", hash)
	}
}

// ForceFlush executes flushFunc and clears out pending items
func (q *Queue) ForceFlush() {
	if !q.stopped {
		q.forceFlushC <- true
	} else {
		q.l.Error("queue is stopped, can not flush")
	}
}

// Run maintains the queue and executes flushes as necessary
func (q *Queue) Run() {
	for {
		select {
		case <-q.stopC:
			println("stopping")
			q.stop()
			return

		case hash := <-q.pendingC:
			go q.queue(hash)

		case <-time.Tick(q.rate):
			go q.flush()

		case <-q.forceFlushC:
			go q.flush()
		}
	}
}

// Close stops the queue runner and releases queue assets
func (q *Queue) Close() { q.stopC <- true }

func (q *Queue) queue(hash string) {
	q.l.Infow("queueing for indexing",
		"hash", hash)
	println("queing")
	q.fmux.RLock()
	atomic.AddInt32(q.pending, 1)
	q.fmux.RUnlock()
}

func (q *Queue) flush() {
	println("checking if should flush")
	if pending := atomic.LoadInt32(q.pending); pending > q.batchSize {
		println("waiting to flush", pending)
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

	// prevent further entries and release resources
	q.stopped = true
	close(q.forceFlushC)
	close(q.pendingC)

	q.fmux.Unlock()
}
