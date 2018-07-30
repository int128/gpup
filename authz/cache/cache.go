package cache

import (
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
)

// Cache represents a token cache.
type Cache struct {
	Filename string
	Secret   Secret
}

// Secret represents a secret for cache encryption.
type Secret []byte

// New returns a Cache.
// This expands the filename if it contains the home directory mark `~`.
func New(filename string, secret Secret) (*Cache, error) {
	f, err := homedir.Expand(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not determine cache filename: %s", err)
	}
	return &Cache{f, secret}, nil
}

// Get returns the token if the cache file exists.
func (c *Cache) Get() (*oauth2.Token, error) {
	file, err := os.Open(c.Filename)
	if err != nil {
		return nil, fmt.Errorf("Could not open cache file %s: %s", c.Filename, err)
	}
	defer file.Close()
	log.Printf("Using token cache file %s", c.Filename)
	return read(file, c.Secret)
}

// Create stores token into a cache file.
func (c *Cache) Create(token *oauth2.Token) error {
	file, err := os.OpenFile(c.Filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Could not create cache file %s: %s", c.Filename, err)
	}
	defer file.Close()
	log.Printf("Storing token cache to %s", c.Filename)
	return write(file, token, c.Secret)
}
