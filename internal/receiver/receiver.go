package receiver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/internal/handlers"
	"github.com/markonick/gigs-challenge/internal/utils"
)

type Receiver struct {
	handler *handlers.Handler
}

func NewReceiver(handler *handlers.Handler) *Receiver {
	return &Receiver{
		handler: handler,
	}
}

func (r *Receiver) HandleGigsEvent(c *gin.Context) {
	gigsEvent, err := ParsePubSubMessage(c)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err, "Failed to parse Pub/Sub message")
		return
	}

	if err := r.handler.ProcessEvent(c.Request.Context(), gigsEvent); err != nil {
		utils.HandleError(c, http.StatusInternalServerError, err, "Failed to handle event")
		return
	}

	c.Status(http.StatusAccepted) // 202 Accepted
}
