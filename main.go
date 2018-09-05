package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/int128/gpup/authz"
	"github.com/int128/gpup/authz/cache"
	"github.com/int128/gpup/debug"
	"github.com/int128/gpup/photos"
	flags "github.com/jessevdk/go-flags"
	"golang.org/x/oauth2"
)

// Set by goreleaser, see https://goreleaser.com/environment/
var version = "1.x"

type options struct {
	NewAlbum           string `short:"n" long:"new-album" value-name:"TITLE" description:"Create an album and add files into it"`
	OAuthMethod        string `long:"oauth-method" default:"browser" choice:"browser" choice:"cli" description:"OAuth authorization method"`
	OAuthCacheFilename string `long:"oauth-cache-filename" default:"~/.gpup_token" description:"OAuth token cache filename"`
	ClientID           string `long:"google-client-id" env:"GOOGLE_CLIENT_ID" required:"1" description:"Google API client ID"`
	ClientSecret       string `long:"google-client-secret" env:"GOOGLE_CLIENT_SECRET" required:"1" description:"Google API client secret"`
	Debug              bool   `long:"debug" env:"DEBUG" description:"Enable request and response logging"`
}

func (o *options) authzConfig() oauth2.Config {
	return oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		Endpoint:     photos.Endpoint,
		Scopes:       photos.Scopes,
	}
}

func (o *options) authzFlow() authz.Flow {
	switch o.OAuthMethod {
	case "browser":
		return &authz.BrowserAuthCodeFlow{Config: o.authzConfig(), Port: 8000}
	case "cli":
		return &authz.CLIAuthCodeFlow{Config: o.authzConfig()}
	default:
		log.Fatalf("Invalid oauth-method: %s", o.OAuthMethod)
		return nil
	}
}

func parseOptions() (*options, []string, error) {
	var o options
	parser := flags.NewParser(&o, flags.HelpFlag)
	parser.Usage = "[OPTIONS] FILE or DIRECTORY..."
	parser.LongDescription = fmt.Sprintf(`
		Version %s

		Setup:
		1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
		2. Enable Photos Library API.
		3. Open https://console.cloud.google.com/apis/credentials
		4. Create an OAuth client ID where the application type is other.
		5. Export GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET variables or set the options.`,
		version)
	args, err := parser.Parse()
	if err != nil {
		return nil, nil, err
	}
	if len(args) == 0 {
		return nil, nil, fmt.Errorf("Too few argument")
	}
	return &o, args, nil
}

func main() {
	opts, args, err := parseOptions()
	if err != nil {
		log.Fatal(err)
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

	tokenCache, err := cache.New(opts.OAuthCacheFilename, cache.Secret(opts.ClientID+opts.ClientSecret))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	token, err := authz.GetToken(ctx, opts.authzFlow(), tokenCache)
	if err != nil {
		log.Fatal(err)
	}
	config := opts.authzConfig()
	client := config.Client(ctx, token)
	if err != nil {
		log.Fatal(err)
	}
	if opts.Debug {
		client = debug.NewClient(client)
	}

	service, err := photos.New(client)
	if err != nil {
		log.Fatal(err)
	}
	if opts.NewAlbum != "" {
		_, err := service.CreateAlbum(ctx, opts.NewAlbum, files)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if err := service.AddToLibrary(ctx, files); err != nil {
			log.Fatal(err)
		}
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
