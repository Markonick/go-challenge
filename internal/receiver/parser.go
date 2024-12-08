package receiver

import (
	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/internal/models"
)

func ParsePubSubMessage(c *gin.Context) (models.BaseEvent, error) {
	// 1. Parse the Pub/Sub message from the request body
	// var pubsubMessage models.PubSubMessage
	var gigsEvent models.BaseEvent
	if err := c.ShouldBindJSON(&gigsEvent); err != nil {
		return models.BaseEvent{}, err
	}

	// // 2. Base64 decode the data field
	// decoded, err := base64.StdEncoding.DecodeString(pubsubMessage.Message.Data)
	// if err != nil {
	// 	return models.BaseEvent{}, err
	// }

	// // 3. Parse the decoded JSON data into our BaseEvent
	// var gigsEvent models.BaseEvent
	// if err := json.Unmarshal(decoded, &gigsEvent); err != nil {
	// 	return models.BaseEvent{}, err
	// }

	return gigsEvent, nil
}
