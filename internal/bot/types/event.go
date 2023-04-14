package types

type EventType string

const (
	CommandAdded   EventType = "comm_added"
	CommandEdited  EventType = "comm_edited"
	CommandDeleted EventType = "comm_deleted"
	CommandCalled  EventType = "comm_called"

	UserTimeout EventType = "user_timeout"
	UserBan     EventType = "user_ban"
)

type CommandEvent struct {
	Issuer      string `json:"issuer"`
	CommandName string `json:"command_name"`
}

type UserEvent struct {
	Target   string `json:"target"`
	Duration int    `json:"duration"`
}
