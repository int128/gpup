package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
	yaml "gopkg.in/yaml.v2"
)

func readConfig(name string, c *externalConfig) error {
	p, err := homedir.Expand(name)
	if err != nil {
		return fmt.Errorf("Could not expand %s: %s", name, err)
	}
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("Could not open %s: %s", name, err)
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	if err := d.Decode(&c); err != nil {
		return fmt.Errorf("Could not read YAML: %s", err)
	}
	return nil
}

func writeConfig(name string, c *externalConfig) error {
	p, err := homedir.Expand(name)
	if err != nil {
		return fmt.Errorf("Could not expand %s: %s", name, err)
	}
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Could not open %s: %s", name, err)
	}
	defer f.Close()
	e := yaml.NewEncoder(f)
	if err := e.Encode(c); err != nil {
		return fmt.Errorf("Could not write to YAML: %s", err)
	}
	return nil
}

// EncodedToken is a base64 encoded json of token.
type EncodedToken string

// Decode returns the token object.
func (t EncodedToken) Decode() (*oauth2.Token, error) {
	if t == "" {
		return nil, nil
	}
	b, err := base64.StdEncoding.DecodeString(string(t))
	if err != nil {
		return nil, fmt.Errorf("Invalid base64: %s", err)
	}
	var token oauth2.Token
	if err := json.Unmarshal(b, &token); err != nil {
		return nil, fmt.Errorf("Invalid json: %s", err)
	}
	return &token, nil
}

// EncodeToken returns an EncodedToken.
func EncodeToken(token *oauth2.Token) (EncodedToken, error) {
	b, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("Could not encode: %s", err)
	}
	return EncodedToken(base64.StdEncoding.EncodeToString(b)), nil
}
