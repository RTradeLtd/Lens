package queue

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type Item struct {
	Key string
	Val interface{}
}

// Queue handles document indexing
type Queue struct {
	l *zap.SugaredLogger

	mux       sync.RWMutex
	closeFunc func() error
	flushFunc func([]*Item) error

	pendingC     chan *Item
	pendingItems []*Item
	pending      int
	rate         time.Duration
	batchSize    int

	stopC   chan bool
	stopped bool
}

// Options declares queue mangement options
type Options struct {
	Rate      time.Duration
	BatchSize int
}

// New instantiates a new queue. flushFunc is is used for periodic index flusing,
// and closeFunc will be used when closing. flushFunc should add items with values,
// and delete items without values. Nil items are possible.
func New(
	logger *zap.SugaredLogger,
	flushFunc func([]*Item) error,
	closeFunc func() error,
	opts Options,
) *Queue {
	if flushFunc == nil {
		flushFunc = func([]*Item) error { return nil }
	}
	if closeFunc == nil {
		closeFunc = func() error { return nil }
	}
	if opts.Rate == 0 {
		opts.Rate = 5 * time.Second
	}

	return &Queue{
		l: logger,

		closeFunc: closeFunc,
		flushFunc: flushFunc,

		pendingC:     make(chan *Item, 3*opts.BatchSize),
		pending:      0,
		pendingItems: make([]*Item, opts.BatchSize),
		rate:         opts.Rate,
		batchSize:    opts.BatchSize,

		stopC:   make(chan bool, 1),
		stopped: false,
	}
}

// Queue indicates that a new item is pending insertion. A nil value indicates
// the item should be deleted.
func (q *Queue) Queue(item *Item) {
	if !q.stopped {
		q.pendingC <- item
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
			q.flush()

		case item := <-q.pendingC:
			q.pendingItems[q.pending] = item
			q.pending++

		case <-q.stopC:
			q.l.Infow("stopping background job")
			ticker.Stop()
			return
		}
	}
}

// Close stops the queue runner and releases queue assets
func (q *Queue) Close() {
	q.stopC <- true
	q.stop()
}

func (q *Queue) flush() {
	if q.pending > q.batchSize {
		q.l.Infow("executing flush", "items", q.pending)
		var now = time.Now()
		q.flushFunc(q.pendingItems)
		q.pending = 0
		q.pendingItems = make([]*Item, q.batchSize)
		q.l.Infow("flush complete",
			"items", q.pending,
			"duration", time.Since(now))
	}
}

func (q *Queue) stop() {
	q.l.Infow("executing close",
		"items", q.pending)
	var now = time.Now()
	q.flushFunc(q.pendingItems)
	q.closeFunc()
	q.l.Infow("flush complete",
		"items", q.pending,
		"duration", time.Since(now))

	// prevent further entries
	q.stopped = true
}
