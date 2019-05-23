package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"

	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// GenerateRSAKeyPair generates an *rsa.PrivateKey and returns the private key
// along a PEM encoded copy of the public key
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return key, EncodePubKeyToPEM(&key.PublicKey), nil
}

// DecodePubKeyPEM decodes a PEM encoded public key to an *rsa.PublicKey
func DecodePubKeyPEM(pk []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pk)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// EncodePubKeyToPEM encodes an *rsa.PublicKey onto a PEM block
func EncodePubKeyToPEM(pk *rsa.PublicKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(pk),
	})
}

// EncryptMessage encrypts a plaintext message with a public key
func EncryptMessage(plaintxt []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha512.New()
	cyphertxt, err := rsa.EncryptOAEP(hash, rand.Reader, pub, plaintxt, nil)
	if err != nil {
		return nil, err
	}
	return cyphertxt, nil
}

// DecryptMessage decrypts an encrypted message with a private key
func DecryptMessage(cyphertxt []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha512.New()
	plaintxt, err := rsa.DecryptOAEP(hash, rand.Reader, priv, cyphertxt, nil)
	if err != nil {
		return nil, err
	}
	return plaintxt, nil
}
