package oauth

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/http"
)

// NewClient creates an OAuth client by given method.
func NewClient(ctx context.Context, method string, clientID string, clientSecret string) (*http.Client, error) {
	switch method {
	case "browser":
		config := NewConfigForBrowser(clientID, clientSecret)
		token, err := GetTokenFromCacheOrServer(ctx, config, GetTokenViaBrowser)
		if err != nil {
			return nil, fmt.Errorf("Could not create oauth client: %s", err)
		}
		return config.Client(ctx, token), nil
	case "cli":
		config := NewConfigForCLI(clientID, clientSecret)
		token, err := GetTokenFromCacheOrServer(ctx, config, GetTokenViaCLI)
		if err != nil {
			return nil, fmt.Errorf("Could not create oauth client: %s", err)
		}
		return config.Client(ctx, token), nil
	default:
		return nil, fmt.Errorf("Unknown oauth method")
	}
}

func generateOAuthState() (string, error) {
	var n uint64
	if err := binary.Read(rand.Reader, binary.LittleEndian, &n); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", n), nil
}
