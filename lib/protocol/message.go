package protocol

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"

	"github.com/adrianosela/botnet/lib/encryption"
)

// Message is a slave-to-master message or response
type Message struct {
	BotID string      `json:"bot_id,omitempty"`
	ReqID string      `json:"req_id,omitempty"`
	Type  MessageType `json:"type"`
	Args  MessageArgs `json:"args"`
}

// MessageType is the type of message being sent to the master
type MessageType string

// MessageArgs is an abstraction for message arguments
type MessageArgs map[string]string

var (
	// MessageTypeJoin is the message type for a JOIN request
	MessageTypeJoin = MessageType("JOIN")
	// MessageTypePong is the response message type for a PING command
	MessageTypePong = MessageType("PONG")
	// MessageTypeSysInfo is the response message type for a SYS_INFO command
	MessageTypeSysInfo = MessageType("SYS_INFO")
	/*
	 * add message types here
	 */
)

const (
	// JoinArgBotPubKey is the argument name for the public key
	// this is only used in messages of MessageTypeJoin
	JoinArgBotPubKey = "public-key"

	// SysInfoArgHoststat TODO
	SysInfoArgHoststat = "host-stat"
)

// Encrypt converts a message to JSON and then encrypts it, returning the bytes
func (m *Message) Encrypt(pub *rsa.PublicKey) ([]byte, error) {
	msgJSON, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("could not marshal message: %s", err)
	}
	msgEncr, err := encryption.EncryptMessage(msgJSON, pub)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt message: %s", err)
	}
	return msgEncr, nil
}

// DecryptMessage  decrypts a JSON message, and then unmarshals it onto a Message type
func DecryptMessage(m []byte, priv *rsa.PrivateKey) (*Message, error) {
	msgJSON, err := encryption.DecryptMessage(m, priv)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt command: %s", err)
	}
	var msg Message
	if err = json.Unmarshal(msgJSON, &msg); err != nil {
		return nil, fmt.Errorf("could not unmarshal command: %s", err)
	}
	return &msg, nil
}
