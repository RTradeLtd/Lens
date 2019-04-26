package queue

import (
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	var q = New(zaptest.NewLogger(t).Sugar(), nil, nil, Options{})
	if q == nil {
		t.Error("got nil")
	}
}

func TestQueue_Queue(t *testing.T) {
	type args struct {
		item *Item
	}
	tests := []struct {
		name       string
		args       args
		wantClosed bool
		wantErr    bool
	}{
		{"invalid item", args{nil}, false, true},
		{"closed queue", args{&Item{Key: "asdf"}}, true, true},
		{"ok", args{&Item{Key: "asdf"}}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var q = New(zaptest.NewLogger(t).Sugar(), nil, nil, Options{
				Rate:      500 * time.Millisecond,
				BatchSize: 1,
			})
			go q.Run()
			if tt.wantClosed {
				q.Close()
			}
			time.Sleep(500 * time.Millisecond)
			if err := q.Queue(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("Queue.Queue() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(time.Second)
			q.Close()
		})
	}
}

func TestQueue_IsStopped(t *testing.T) {
	var q = New(zaptest.NewLogger(t).Sugar(), nil, nil, Options{
		Rate:      500 * time.Millisecond,
		BatchSize: 1,
	})
	if b := q.IsStopped(); b != q.stopped {
		t.Errorf("IsStopped = '%v', got '%v'", b, q.stopped)
	}
}
