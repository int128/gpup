package oauth

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestReadWriteTokenCache(t *testing.T) {
	config := &oauth2.Config{
		ClientID:     "P0PvRC7uASR0ZryVJStYfIXMmOHiZBOyW1noaroIVq2ty0myCz21O4yZFmJY1qK1dl5klqti",
		ClientSecret: "BwmhGP0f78io9cKLe64gc2dd",
	}
	token := &oauth2.Token{
		AccessToken:  "ZybjMlBaImyfzugt0CeYsl51BvFPZQgPw7u1oPIDQ0EZ88qlS12Wp9yk6PX86VCK",
		RefreshToken: "wXanOiKsVEClhRInZ2VIm0gohR7FdGIdLgK8gI9iBxJSbjUGBvFfkEy5T4aHzKQy",
		Expiry:       time.Date(2018, 6, 1, 2, 3, 4, 0, time.UTC),
		TokenType:    "Bearer",
	}

	f, err := ioutil.TempFile("", "TokenCache")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	if err := WriteTokenCache(f, token, config); err != nil {
		t.Fatal(err)
	}

	fs, err := os.Stat(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Created %d bytes file", fs.Size())
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	actual, err := ReadTokenCache(f, config)
	if err != nil {
		t.Fatal(err)
	}

	if token.AccessToken != actual.AccessToken {
		t.Errorf("AccessToken: wants %s but %s", token.AccessToken, actual.AccessToken)
	}
	if token.RefreshToken != actual.RefreshToken {
		t.Errorf("RefreshToken: wants %s but %s", token.RefreshToken, actual.RefreshToken)
	}
	if token.Expiry != actual.Expiry {
		t.Errorf("Expiry: wants %v but %v", token.Expiry, actual.Expiry)
	}
	if token.TokenType != actual.TokenType {
		t.Errorf("TokenType: wants %s but %s", token.TokenType, actual.TokenType)
	}
}
