package models

// PubSubMessage is the wrapper for the Pub/Sub message
type PubSubMessage struct {
	Message struct {
		Data string `json:"data"`
	} `json:"message"`
}
