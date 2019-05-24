package protocol

// Message is a slave-to-master message or response
type Message struct {
	ReqID string      `json:"omitempty,id"`
	Type  MessageType `json:"type"`
	Args  MessageArgs `json:"args"`
}

// MessageType is the type of message being sent to the master
type MessageType string

// MessageArgs is an abstraction for message arguments
type MessageArgs map[string]string

var (
	// MessageTypeJoin is a the message type for a join request
	MessageTypeJoin = MessageType("JOIN")
	/*
	 * add message types here
	 */
)

const (
	// JoinArgBotPubKey is the argument name for the public key
	// this is only used in messages of MessageTypeJoin
	JoinArgBotPubKey = "public-key"
)
