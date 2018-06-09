package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/int128/gpup/oauth"
	"github.com/int128/gpup/photos"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
}

func main() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		printOAuthConfigError()
		os.Exit(1)
	}

	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "[OPTIONS] DIRECTORIES..."
	args, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	if len(args) == 0 {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := oauth.NewClient(ctx, clientID, clientSecret)
	if err != nil {
		log.Fatal(err)
	}
	service, err := photos.New(client)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: find files in the directory
	album, err := service.CreateAlbum("Test", args)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Successfuly created an album %s", album.Title)
}

func printOAuthConfigError() {
	fmt.Print(`Error: GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set.
--------
Follow the steps:
1. Open https://console.cloud.google.com/apis/credentials
2. Create an OAuth client ID where the application type is other.
3. Set the following environment variables:
export GOOGLE_CLIENT_ID=
export GOOGLE_CLIENT_SECRET=
--------
`)
}
