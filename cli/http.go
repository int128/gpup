package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/int128/gpup/photos"
	"github.com/int128/oauth2cli"
	"golang.org/x/oauth2"
)

func (c *CLI) newHTTPClient() *http.Client {
	if c.Debug {
		return &http.Client{Transport: loggingTransport{http.DefaultTransport}}
	}
	return http.DefaultClient
}

func (c *CLI) newOAuth2Client(ctx context.Context) (*http.Client, error) {
	token, err := c.ExternalConfig.EncodedToken.Decode()
	if err != nil {
		return nil, fmt.Errorf("Invalid config: %s", err)
	}
	oauth2Config := oauth2.Config{
		ClientID:     c.ExternalConfig.ClientID,
		ClientSecret: c.ExternalConfig.ClientSecret,
		Endpoint:     photos.Endpoint,
		Scopes:       photos.Scopes,
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, c.newHTTPClient())
	switch {
	case token == nil:
		flow := oauth2cli.AuthCodeFlow{Config: oauth2Config}
		token, err = flow.GetToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("Could not get a token: %s", err)
		}
		c.ExternalConfig.EncodedToken, err = EncodeToken(token)
		if err != nil {
			return nil, fmt.Errorf("Could not encode the token: %s", err)
		}
		if err := c.ExternalConfig.Write(c.ConfigName); err != nil {
			return nil, fmt.Errorf("Could not write the token to %s: %s", c.ConfigName, err)
		}
		log.Printf("Saved token to %s", c.ConfigName)

	case !token.Valid():
		log.Printf("Token of %s has been expired, refreshing", c.ConfigName)
		token, err = oauth2Config.TokenSource(ctx, token).Token()
		if err != nil {
			return nil, fmt.Errorf("Could not refresh the token: %s", err)
		}
		c.ExternalConfig.EncodedToken, err = EncodeToken(token)
		if err != nil {
			return nil, fmt.Errorf("Could not encode the token: %s", err)
		}
		if err := c.ExternalConfig.Write(c.ConfigName); err != nil {
			return nil, fmt.Errorf("Could not write token to %s: %s", c.ConfigName, err)
		}
		log.Printf("Saved token to %s", c.ConfigName)
	}
	client := oauth2Config.Client(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("Could not create a client: %s", err)
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
