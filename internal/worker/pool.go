package worker

import (
	"context"

	"github.com/gammazero/workerpool"
	"github.com/markonick/gigs-challenge/internal/logger"
)

// Task represents a unit of work to be processed by the worker pool.
type Task interface {
	Process(ctx context.Context) error
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

	logger.Log.Info().
		Str("task_id", task.ID()).
		Msg("Submitting task to worker pool")

	p.wp.Submit(func() {
		logger.Log.Info().
			Str("task_id", task.ID()).
			Msg("Starting task processing")

		if err := task.Process(ctx); err != nil {
			logger.Log.Error().
				Err(err).
				Str("task_id", task.ID()).
				Msg("Failed to process task")
			return
		}

		logger.Log.Info().
			Str("task_id", task.ID()).
			Msg("Task processed successfully")
	})

	return nil
}

func (p *Pool) Close() {
	p.wp.StopWait()
}
