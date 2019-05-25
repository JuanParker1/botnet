package protocol

// Command is a master-to-slave instruction
type Command struct {
	ReqID  string      `json:"id"`
	Type   CommandType `json:"type"`
	Args   CommandArgs `json:"args"`
	Target string      `json:"target"`
	When   int64       `json:"runat"`
}

// CommandArgs is an abstraction for command arguments
type CommandArgs map[string]string

// CommandType corresponds to a desired action being requested from slaves
type CommandType string

var (
	// CommandTypeWelcome is the response to a successful botnet join attempt
	// from a slave
	CommandTypeWelcome = CommandType("WELCOME")
	// CommandTypePing requests a PONG from bots, allowing calculations of latency
	CommandTypePing = CommandType("PING")
	/*
	 * add command types here
	 */
)
