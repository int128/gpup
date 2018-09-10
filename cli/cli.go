package cli

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/int128/gpup/authz"

	"github.com/int128/gpup/debug"
	"github.com/int128/gpup/photos"
	flags "github.com/jessevdk/go-flags"
	"golang.org/x/oauth2"
)

// CLI has the command options.
type CLI struct {
	NewAlbum     string `short:"n" long:"new-album" value-name:"TITLE" description:"Create an album and add files into it"`
	ClientID     string `long:"google-client-id" env:"GOOGLE_CLIENT_ID" required:"1" description:"Google API client ID"`
	ClientSecret string `long:"google-client-secret" env:"GOOGLE_CLIENT_SECRET" required:"1" description:"Google API client secret"`
	Debug        bool   `long:"debug" env:"DEBUG" description:"Enable request and response logging"`
	paths        []string
}

// Parse command line and returns a CLI.
func Parse(osArgs []string, version string) (*CLI, error) {
	var o CLI
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
	args, err := parser.ParseArgs(osArgs[1:])
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("Too few argument")
	}
	o.paths = args
	return &o, nil
}

// Run runs the command.
func (c *CLI) Run() error {
	files, err := findFiles(c.paths)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("File not found in %s", strings.Join(c.paths, ", "))
	}
	log.Printf("The following %d files will be uploaded:", len(files))
	for i, file := range files {
		fmt.Printf("%3d: %s\n", i+1, file)
	}

	ctx := context.Background()
	oauth2Config := oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     photos.Endpoint,
		Scopes:       photos.Scopes,
		RedirectURL:  "http://localhost:8000",
	}
	flow := authz.AuthCodeFlow{
		Config:     &oauth2Config,
		ServerPort: 8000,
	}
	token, err := flow.GetToken(ctx)
	if err != nil {
		return err
	}
	client := oauth2Config.Client(ctx, token)
	if err != nil {
		return err
	}
	if c.Debug {
		client = debug.NewClient(client)
	}

	service, err := photos.New(client)
	if err != nil {
		return err
	}
	if c.NewAlbum != "" {
		_, err := service.CreateAlbum(ctx, c.NewAlbum, files)
		if err != nil {
			return err
		}
	} else {
		if err := service.AddToLibrary(ctx, files); err != nil {
			return err
		}
	}
	return nil
}
