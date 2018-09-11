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

// CLI represents input for the command.
type CLI struct {
	ConfigName string `long:"gpupconfig" env:"GPUPCONFIG" default:"~/.gpupconfig" description:"Path to the config file"`
	NewAlbum   string `short:"n" long:"new-album" value-name:"TITLE" description:"Create an album and add files into it"`
	Debug      bool   `long:"debug" env:"DEBUG" description:"Enable request and response logging"`

	externalConfig

	Paths []string
}

type externalConfig struct {
	ClientID     string `yaml:"client-id" long:"google-client-id" env:"GOOGLE_CLIENT_ID" description:"Google API client ID"`
	ClientSecret string `yaml:"client-secret" long:"google-client-secret" env:"GOOGLE_CLIENT_SECRET" description:"Google API client secret"`
	EncodedToken string `yaml:"token" long:"google-token" env:"GOOGLE_TOKEN" description:"Google API token (base64 encoded json)"`
}

// New parses the arguments, read the config and returns a CLI.
func New(osArgs []string, version string) (*CLI, error) {
	var o CLI
	parser := flags.NewParser(&o, flags.HelpFlag)
	parser.Usage = "[OPTIONS] FILE or DIRECTORY..."
	parser.LongDescription = fmt.Sprintf("Version %s", version)
	if _, err := parser.ParseArgs(osArgs[1:]); err != nil {
		return nil, err
	}
	if err := readConfig(o.ConfigName, &o.externalConfig); err != nil {
		return nil, fmt.Errorf("Could not read config: %s", err)
	}
	args, err := parser.ParseArgs(osArgs[1:])
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("Too few argument")
	}
	o.Paths = args
	return &o, nil
}

// Run runs the command.
func (c *CLI) Run() error {
	files, err := findFiles(c.Paths)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("File not found in %s", strings.Join(c.Paths, ", "))
	}
	log.Printf("The following %d files will be uploaded:", len(files))
	for i, file := range files {
		fmt.Printf("%3d: %s\n", i+1, file)
	}

	ctx := context.Background()
	token, err := c.GetToken()
	if err != nil {
		return fmt.Errorf("Invalid config: %s", err)
	}
	oauth2Config := oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     photos.Endpoint,
		Scopes:       photos.Scopes,
		RedirectURL:  "http://localhost:8000",
	}
	if token == nil {
		flow := authz.AuthCodeFlow{
			Config:     &oauth2Config,
			ServerPort: 8000,
		}
		token, err = flow.GetToken(ctx)
		if err != nil {
			return err
		}
		c.SetToken(token)
		if err := writeConfig(c.ConfigName, &c.externalConfig); err != nil {
			return fmt.Errorf("Could not write token to %s: %s", c.ConfigName, err)
		}
		log.Printf("Saved token to %s", c.ConfigName)
	} else {
		log.Printf("Using token in %s", c.ConfigName)
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
