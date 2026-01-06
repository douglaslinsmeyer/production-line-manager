package mqtt

// MQTT topic constants
const (
	// Production line event topics (published by API)
	TopicEventCreated = "production-lines/events/created"
	TopicEventUpdated = "production-lines/events/updated"
	TopicEventDeleted = "production-lines/events/deleted"
	TopicEventStatus  = "production-lines/events/status"

	// Production line command topics (subscribed by API)
	TopicCommandStatus = "production-lines/commands/status"

	// Signal profile event topics (published by API)
	TopicProfileCreated  = "signal-profiles/events/created"
	TopicProfileUpdated  = "signal-profiles/events/updated"
	TopicProfileDeleted  = "signal-profiles/events/deleted"
	TopicProfileAssigned = "signal-profiles/events/assigned"

	// Device command topics (published by API)
	// Format: devices/{mac}/command
	// Individual device commands are formatted at runtime
)
