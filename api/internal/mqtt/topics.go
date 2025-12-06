package mqtt

// MQTT topic constants
const (
	// Event topics (published by API)
	TopicEventCreated = "production-lines/events/created"
	TopicEventUpdated = "production-lines/events/updated"
	TopicEventDeleted = "production-lines/events/deleted"
	TopicEventStatus  = "production-lines/events/status"

	// Command topics (subscribed by API)
	TopicCommandStatus = "production-lines/commands/status"
)
