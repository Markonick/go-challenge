package models

// EventType represents a Gigs webhook event type
type EventType string

const (
	SubscriptionUpdated   EventType = "subscription.updated"
	SubscriptionRenewed   EventType = "subscription.renewed"
	PaymentSucceeded      EventType = "payment.succeeded"
	TaxCreated            EventType = "tax.created"
	TaxRateUpdated        EventType = "taxRate.updated"
	UserAddressCreated    EventType = "user.address.created"
	UserCreated           EventType = "user.created"
	SubscriptionActivated EventType = "subscription.activated"
	SubscriptionCanceled  EventType = "subscription.canceled"
	SubscriptionCreated   EventType = "subscription.created"
	SubscriptionEnded     EventType = "subscription.ended"
	UserUpdated           EventType = "user.updated"
	UserAddressUpdated    EventType = "user.address.updated"
	UserAddressRenewed    EventType = "user.address.renewed"
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
		UserAddressCreated,
		UserAddressUpdated,
		UserAddressRenewed,
		TaxRateUpdated,
		TaxCreated,
		PaymentSucceeded,
	}
}
