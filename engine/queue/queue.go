package queue

import (
	"errors"
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

		pendingC:     make(chan *Item, opts.BatchSize),
		pending:      0,
		pendingItems: make([]*Item, opts.BatchSize),
		rate:         opts.Rate,
		batchSize:    opts.BatchSize,

		stopC:   make(chan bool, 1),
		stopped: true,
	}
}

// Queue indicates that a new item is pending insertion. A nil value indicates
// the item should be deleted.
func (q *Queue) Queue(item *Item) error {
	if !q.stopped {
		q.pendingC <- item
	} else {
		q.l.Error("queue failed: queue is stopped, cannot queue more elements")
		return errors.New("queue is stopped, cannot queue more element")
	}
	return nil
}

// Run maintains the queue and executes flushes as necessary
func (q *Queue) Run() {
	q.stopped = false
	q.l.Infow("spinning up queue", "rate", q.rate)
	var ticker = time.NewTicker(q.rate)
	for {
		select {
		case <-ticker.C:
			q.flushIfNeeded()

		case item := <-q.pendingC:
			q.pendingItems[q.pending] = item
			q.pending++
			q.flushIfNeeded()

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

func (q *Queue) IsStopped() bool { return q.stopped }

func (q *Queue) flushIfNeeded() {
	if q.pending >= q.batchSize {
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
	if err := q.flushFunc(q.pendingItems); err != nil {
		q.l.Errorw("unable to flush", "error", err)
	}
	if err := q.closeFunc(); err != nil {
		q.l.Errorw("error occured on close", "error", err)
	}
	q.l.Infow("queue and index closed",
		"duration", time.Since(now))

	// prevent further entries
	q.stopped = true
}
