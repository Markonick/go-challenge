package worker

import (
	"context"

	"github.com/gammazero/workerpool"
)

// Task represents a unit of work to be processed by the worker pool.
type Task interface {
	Execute(ctx context.Context) error
	ID() string
}

// Pool that manages concurrent task processing
type Pool struct {
	wp *workerpool.WorkerPool
}

func NewPool(maxWorkers int) *Pool {
	return &Pool{
		wp: workerpool.New(maxWorkers),
	}
}

func (p *Pool) ProcessTask(task Task) error {
	// Create a new background context for the task
	ctx := context.Background()
	errChan := make(chan error, 1)

	p.wp.Submit(func() {
		err := task.Execute(ctx)
		errChan <- err // Send the error (or nil)
		close(errChan) // Always close the channel
	})

	return <-errChan // Wait for result
}

func (p *Pool) Close() {
	p.wp.StopWait()
}
