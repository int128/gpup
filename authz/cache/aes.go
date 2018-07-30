package cache

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/oauth2"
)

func read(r io.Reader, secret Secret) (*oauth2.Token, error) {
	encrypted, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Could not read token cache: %s", err)
	}
	stream, err := newAESCTR(secret)
	if err != nil {
		return nil, fmt.Errorf("Could not initialize cipher: %s", err)
	}
	tokenJSON := make([]byte, len(encrypted))
	stream.XORKeyStream(tokenJSON, encrypted)
	var token oauth2.Token
	if err := json.Unmarshal(tokenJSON, &token); err != nil {
		return nil, fmt.Errorf("Could not decode oauth2 token: %s: %v", err, tokenJSON)
	}
	return &token, nil
}

func write(w io.Writer, token *oauth2.Token, secret Secret) error {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("Could not encode oauth2 token: %s", err)
	}
	stream, err := newAESCTR(secret)
	if err != nil {
		return fmt.Errorf("Could not initialize cipher: %s", err)
	}
	encrypted := make([]byte, len(tokenJSON))
	stream.XORKeyStream(encrypted, tokenJSON)
	n, err := w.Write(encrypted)
	if err != nil {
		return fmt.Errorf("Could not write token cache: %s", err)
	}
	if n != len(encrypted) {
		return fmt.Errorf("Could not write full token cache: wants %d but %d bytes", len(encrypted), n)
	}
	return nil
}

func newAESCTR(secret Secret) (cipher.Stream, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("Too short secret: %d", len(secret))
	}
	key := secret[0:32]
	iv := secret[len(secret)-aes.BlockSize : len(secret)]
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("Could not create AES cipher: %s", err)
	}
	return cipher.NewCTR(c, iv), nil
}
