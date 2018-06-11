package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/int128/gpup/oauth"
	"github.com/int128/gpup/photos"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	AlbumTitle  string `short:"a" long:"album-title" value-name:"TITLE" description:"Create an album and add files into it"`
	OAuthMethod string `long:"oauth-method" default:"browser" choice:"browser" choice:"cli" description:"OAuth method"`
}

func main() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		printOAuthConfigError()
		os.Exit(1)
	}

	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "[OPTIONS] FILE or DIRECTORY..."
	args, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	if len(args) == 0 {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	files, err := findFiles(args)
	if err != nil {
		log.Fatal(err)
	}
	if len(files) == 0 {
		log.Fatalf("File not found in %s", strings.Join(args, ", "))
	}
	log.Printf("The following %d files will be uploaded:", len(files))
	for i, file := range files {
		fmt.Printf("%3d: %s\n", i+1, file)
	}

	ctx := context.Background()
	client, err := func() (*http.Client, error) {
		switch opts.OAuthMethod {
		case "browser":
			return oauth.NewClientViaBrowser(ctx, clientID, clientSecret)
		case "cli":
			return oauth.NewClientViaCLI(ctx, clientID, clientSecret)
		default:
			return nil, fmt.Errorf("Unknown oauth-method")
		}
	}()
	if err != nil {
		log.Fatal(err)
	}
	service, err := photos.New(client)
	if err != nil {
		log.Fatal(err)
	}

	if opts.AlbumTitle != "" {
		album, err := service.CreateAlbum(opts.AlbumTitle, files)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Successfuly created the album %s", album.Title)
	} else {
		added, err := service.AddToLibrary(files)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Successfuly added %d files to the library", added)
	}
}

func findFiles(filePaths []string) ([]string, error) {
	files := make([]string, 0, len(filePaths)*2)
	for _, parent := range filePaths {
		if err := filepath.Walk(parent, func(child string, info os.FileInfo, err error) error {
			switch {
			case err != nil:
				return err
			case info.Mode().IsRegular():
				files = append(files, child)
				return nil
			default:
				return nil
			}
		}); err != nil {
			return nil, fmt.Errorf("Error while finding files in %s: %s", parent, err)
		}
	}
	return files, nil
}

func printOAuthConfigError() {
	fmt.Print(`Error: GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set.
--------
Follow the steps:
1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
2. Enable Photos Library API.
3. Open https://console.cloud.google.com/apis/credentials
4. Create an OAuth client ID where the application type is other.
5. Set the following environment variables:
export GOOGLE_CLIENT_ID=
export GOOGLE_CLIENT_SECRET=
--------
`)
}
