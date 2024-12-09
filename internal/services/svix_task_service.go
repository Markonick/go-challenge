package services

import (
	"github.com/markonick/gigs-challenge/internal/models"
	"github.com/markonick/gigs-challenge/internal/svix"
	webhook "github.com/markonick/gigs-challenge/internal/webhooks"
	"github.com/markonick/gigs-challenge/internal/worker"
)

type TaskService interface {
	ProcessEvent(event models.BaseEvent) error
}

type taskServiceImpl struct {
	svixClient    svix.Client
	projectAppIDs map[string]string
	workerPool    worker.Pool
}

func NewTaskService(svixClient svix.Client, projectAppIDs map[string]string, workerPool *worker.Pool) TaskService {
	return &taskServiceImpl{
		svixClient:    svixClient,
		projectAppIDs: projectAppIDs,
		workerPool:    *workerPool,
	}
}

func (t *taskServiceImpl) ProcessEvent(event models.BaseEvent) error {
	task := webhook.NewWebhookTask(event, t.svixClient, t.projectAppIDs)
	t.workerPool.ProcessTask(task)
	return nil
}
