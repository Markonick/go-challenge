package receiver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/internal/handlers"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/svix"
	"github.com/markonick/gigs-challenge/internal/utils"
	"github.com/markonick/gigs-challenge/internal/worker"
)

type Receiver struct {
	svixClient    svix.Client
	projectAppIDs map[string]string
	workerPool    *worker.Pool
}

func NewReceiver(svixClient svix.Client, projectAppIDs map[string]string, workerPool *worker.Pool) *Receiver {
	return &Receiver{
		svixClient:    svixClient,
		projectAppIDs: projectAppIDs,
		workerPool:    workerPool,
	}
}

func (r *Receiver) HandleNotification(c *gin.Context) {
	gigsEvent, err := ParsePubSubMessage(c)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err, "Failed to parse Pub/Sub message")
		return
	}

	logger.Log.Info().
		Str("event_type", string(gigsEvent.Type)).
		Str("event_id", gigsEvent.ID).
		Msg("Received event")

	task := handlers.NewWebhookTask(gigsEvent, r.svixClient, r.projectAppIDs)
	r.workerPool.ProcessTask(task)

	logger.Log.Info().
		Str("event_id", gigsEvent.ID).
		Msg("Task submitted to worker pool")

	c.Status(http.StatusAccepted)
}
