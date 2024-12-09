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
	UserAddressCreated    EventType = "user.address.created"
	UserAddressUpdated    EventType = "user.address.updated"
	UserAddressRenewed    EventType = "user.address.renewed"
	TaxRateUpdated        EventType = "tax_rate.updated"
	TaxRateCreated        EventType = "tax_rate.created"
	PaymentSucceeded      EventType = "payment.succeeded"
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
		TaxRateCreated,
		PaymentSucceeded,
	}
}
