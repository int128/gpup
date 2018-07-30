package authz

import (
	"context"
	"fmt"
	"log"

	"github.com/int128/gpup/authz/cache"
	"golang.org/x/oauth2"
)

// Flow represents an authorization method.
type Flow interface {
	GetToken(context.Context) (*oauth2.Token, error)
}

// GetToken returns the token from cache or new one from server.
func GetToken(ctx context.Context, flow Flow, c *cache.Cache) (*oauth2.Token, error) {
	token, err := c.Get()
	if err != nil {
		log.Printf("Proceed authorization due to token cache unavailable: %s", err)
		token, err = flow.GetToken(ctx)
		if err != nil {
			return nil, err
		}
		err = c.Create(token)
		if err != nil {
			return nil, fmt.Errorf("Could not create a token cache: %s", err)
		}
		return token, nil
	}
	return token, nil
}
