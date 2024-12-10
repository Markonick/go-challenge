package worker

import (
	"context"
	"fmt"

	"github.com/gammazero/workerpool"
	"github.com/markonick/gigs-challenge/internal/logger"
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
		if err := task.Execute(ctx); err != nil {
			logger.Log.Error().
				Err(err).
				Str("task_id", task.ID()).
				Str("error_type", fmt.Sprintf("%T", err)).
				Msg("Failed to process task")
			errChan <- err // Send error back
		}

		close(errChan)
		logger.Log.Info().
			Str("task_id", task.ID()).
			Msg("Task processed successfully and closed error channel")
	})

	return <-errChan // Wait for and return the error
}

func (p *Pool) Close() {
	p.wp.StopWait()
}
