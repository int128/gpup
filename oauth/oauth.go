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
		return NewClientViaBrowser(ctx, clientID, clientSecret)
	case "cli":
		return NewClientViaCLI(ctx, clientID, clientSecret)
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
