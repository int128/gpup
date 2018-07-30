package cache

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestReadWriteTokenCache(t *testing.T) {
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
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())
	cache := &Cache{
		Filename: f.Name(),
		Secret:   Secret("FVpZU07PBMy9uIQWnjslNPIbtr82yKyky73ynxOlSLINDKtFiSxkCdkImHXtlY60"),
	}

	if err := cache.Create(token); err != nil {
		t.Fatal(err)
	}
	fs, err := os.Stat(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Created %d bytes file", fs.Size())

	actual, err := cache.Get()
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
