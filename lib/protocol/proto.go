package protocol

import (
	"log"
	"time"
)

// Event is a master-to-slave instruction
type Event struct {
	Type   EventType         `json:"type"`
	Args   map[string]string `json:"args"`
	Target string            `json:"target"`
	When   int64             `json:"runat"`
}

// EventType is the type of commands desired by the master
type EventType string

var (
	// EventTypeLog logs that a message was received
	EventTypeLog = EventType("LOG")

	/*
	 * add event types here
	 */
)

// HandleMasterEvent handles a single event from the master node
func HandleMasterEvent(eve Event) {
	switch eve.Type {
	case EventTypeLog:
		log.Printf("received log event type at %d, full event: %v", time.Now().UnixNano(), eve)
		return
	default:
		log.Printf("received unknown event type at %d, full event: %v", time.Now().UnixNano(), eve)
		return
	}
}
