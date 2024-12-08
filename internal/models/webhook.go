package models

// EventType represents a Gigs webhook event type
type EventType string

const (
	SubscriptionActivated EventType = "subscription.activated"
	SubscriptionCanceled  EventType = "subscription.canceled"
	SubscriptionCreated   EventType = "subscription.created"
	SubscriptionEnded     EventType = "subscription.ended"
	SubscriptionRenewed   EventType = "subscription.renewed"
	SubscriptionUpdated   EventType = "subscription.updated"
	UserCreated           EventType = "user.created"
	UserUpdated           EventType = "user.updated"
)

// GetCommonEventTypes returns all supported webhook event types
func GetCommonEventTypes() []EventType {
	return []EventType{
		SubscriptionActivated,
		SubscriptionCanceled,
		SubscriptionCreated,
		SubscriptionEnded,
		SubscriptionRenewed,
		SubscriptionUpdated,
		UserCreated,
		UserUpdated,
	}
}
