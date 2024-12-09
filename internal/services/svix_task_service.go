package services

import (
	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/worker"
)

type TaskService interface {
	ProcessEvent(event models.BaseEvent) error
}

type taskServiceImpl struct {
	createTask func(models.BaseEvent) worker.Task
	workerPool worker.Pool
}

func NewTaskService(createTask func(models.BaseEvent) worker.Task, workerPool *worker.Pool) TaskService {
	return &taskServiceImpl{
		createTask: createTask,
		workerPool: *workerPool,
	}
}

func (t *taskServiceImpl) ProcessEvent(event models.BaseEvent) error {
	task := t.createTask(event)
	t.workerPool.ProcessTask(task)
	return nil
}
