package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/int128/gpup/authz"
	"github.com/int128/gpup/photos"
	"golang.org/x/oauth2"
)

func (c *CLI) newClient(ctx context.Context) (*http.Client, error) {
	token, err := c.EncodedToken.Decode()
	if err != nil {
		return nil, fmt.Errorf("Invalid config: %s", err)
	}
	oauth2Config := oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     photos.Endpoint,
		Scopes:       photos.Scopes,
		RedirectURL:  "http://localhost:8000",
	}
	if c.Debug {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
			Transport: loggingTransport{http.DefaultTransport},
		})
	}
	if token == nil {
		flow := authz.AuthCodeFlow{
			Config:     &oauth2Config,
			ServerPort: 8000,
		}
		token, err = flow.GetToken(ctx)
		if err != nil {
			return nil, err
		}
		c.EncodedToken, err = EncodeToken(token)
		if err != nil {
			return nil, err
		}
		if err := writeConfig(c.ConfigName, &c.externalConfig); err != nil {
			return nil, fmt.Errorf("Could not write token to %s: %s", c.ConfigName, err)
		}
		log.Printf("Saved token to %s", c.ConfigName)
	} else {
		log.Printf("Using token in %s", c.ConfigName)
	}
	client := oauth2Config.Client(ctx, token)
	if err != nil {
		return nil, err
	}
	if c.Debug {
		client.Transport = loggingTransport{client.Transport}
	}
	return client, nil
}
