package services

import (
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/utils"
	"github.com/markonick/gigs-challenge/internal/worker"
)

type TaskService interface {
	ProcessEvent(event models.BaseEvent) error
}

// Implementation holds channels and task creation function
type taskServiceImpl struct {
	workerPool *worker.Pool
	createTask func(models.BaseEvent) worker.Task
}

func NewTaskService(numWorkers int, createTask func(models.BaseEvent) worker.Task) TaskService {
	logger.Log.Info().
		Int("num_workers", numWorkers).
		Msg("Initializing task service with worker pool")
	return &taskServiceImpl{
		workerPool: worker.NewPool(numWorkers),
		createTask: createTask,
	}
}

func (t *taskServiceImpl) ProcessEvent(event models.BaseEvent) error {
	task := t.createTask(event)
	logger.Log.Info().
		Str("event_id", event.ID).
		Str("task_id", task.ID()).
		Msg("Created task, submitting to worker pool")

	if err := t.workerPool.ProcessTask(task); err != nil {
		// Only log error if it's not a duplicate
		if _, isDuplicate := err.(*utils.ConflictError); !isDuplicate {
			logger.Log.Error().
				Err(err).
				Str("event_id", event.ID).
				Str("task_id", task.ID()).
				Msg("Failed to process task")
		}
		return err
	}

	return nil
}
