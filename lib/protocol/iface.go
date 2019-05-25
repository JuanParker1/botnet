package protocol

import "net/http"

// CC represents all actions that any implementation of a Command and Control
// server should be able to perform
type CC interface {
	// Initialize botnet
	StartBotnet()

	// Register a new bot
	EnrolBot(interface{})

	// Deregister a bot
	ReleaseBot(string)

	// Handle messages with distinct Message.Type values
	HandleBotMessage(*Message)

	// Broadcast a command to all bots in the botnet
	BroadcastCommand(*Command)

	// Send a command to a single bot
	SendCommandToBot(*Command, string)

	// HTTP Handler: Key Discovery
	KeyHTTPHandler(w http.ResponseWriter, r *http.Request)

	// HTTP Handler: Botnet WS
	CommandAndControlHTTPHandler(w http.ResponseWriter, r *http.Request)
}

// Bot represents all actions that any implementation of a Bot should be able
// to perform
type Bot interface {
	// Begin listening to command and control
	Run()

	// Handle a single command from the command and control server
	HandleCommandFromCC(*Command)

	// Send a message to the command and control server, this is typically called
	// in reponse to receiving a command, or upon ancountering an event
	SendMessageToCC(*Message)
}
