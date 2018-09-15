package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/int128/gpup/authz"
	"github.com/int128/gpup/photos"
	"golang.org/x/oauth2"
)

func (c *CLI) newClient(ctx context.Context) (*http.Client, error) {
	token, err := c.ExternalConfig.EncodedToken.Decode()
	if err != nil {
		return nil, fmt.Errorf("Invalid config: %s", err)
	}
	oauth2Config := oauth2.Config{
		ClientID:     c.ExternalConfig.ClientID,
		ClientSecret: c.ExternalConfig.ClientSecret,
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
		c.ExternalConfig.EncodedToken, err = EncodeToken(token)
		if err != nil {
			return nil, err
		}
		if err := c.ExternalConfig.Write(c.ConfigName); err != nil {
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

type loggingTransport struct {
	transport http.RoundTripper
}

func (t loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req != nil {
		dump, err := httputil.DumpRequestOut(req, false)
		if err != nil {
			log.Printf("Could not dump request: %s", err)
		}
		log.Printf("[REQUEST] %s %s\n%s", req.Method, req.URL, string(dump))
	}
	res, err := t.transport.RoundTrip(req)
	if res != nil {
		dump, err := httputil.DumpResponse(res, false)
		if err != nil {
			log.Printf("Could not dump response: %s", err)
		}
		log.Printf("[RESPONSE] %s %s\n%s", req.Method, req.URL, string(dump))
	}
	return res, err
}
