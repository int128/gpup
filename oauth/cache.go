package oauth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
)

// FindTokenCache returns token if the cache file exists.
// Otherwise returns nil.
func FindTokenCache(config *oauth2.Config) (*oauth2.Token, error) {
	name, err := tokenCacheName()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(name)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Could not open cache file %s: %s", name, err)
	}
	defer file.Close()
	log.Printf("Using token cache file %s", name)
	return ReadTokenCache(file, config)
}

// CreateTokenCache stores token into a cache file.
func CreateTokenCache(token *oauth2.Token, config *oauth2.Config) error {
	name, err := tokenCacheName()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Could not create cache file %s: %s", name, err)
	}
	defer file.Close()
	log.Printf("Storing token cache to %s", name)
	return WriteTokenCache(file, token, config)
}

func tokenCacheName() (string, error) {
	name, err := homedir.Expand("~/.gpup_token")
	if err != nil {
		return "", fmt.Errorf("Could not determine cache directory: %s", err)
	}
	return name, nil
}

// ReadTokenCache reads and decrypts the token.
func ReadTokenCache(r io.Reader, config *oauth2.Config) (*oauth2.Token, error) {
	encrypted, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Could not read token cache: %s", err)
	}
	stream, err := newAESCTR(config)
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

// WriteTokenCache encrypts the token and writes.
func WriteTokenCache(w io.Writer, token *oauth2.Token, config *oauth2.Config) error {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("Could not encode oauth2 token: %s", err)
	}
	stream, err := newAESCTR(config)
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

func newAESCTR(config *oauth2.Config) (cipher.Stream, error) {
	secret := []byte(config.ClientID + config.ClientSecret)
	if len(secret) < 32 {
		return nil, fmt.Errorf("Too short ClientID and ClientSecret: %d", len(secret))
	}
	key := secret[0:32]
	iv := secret[len(secret)-aes.BlockSize : len(secret)]
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("Could not create AES cipher: %s", err)
	}
	return cipher.NewCTR(c, iv), nil
}
