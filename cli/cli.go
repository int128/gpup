package cli

import (
	"context"
	"fmt"
	"log"

	flags "github.com/jessevdk/go-flags"
)

// CLI represents input for the command.
type CLI struct {
	ConfigName string `long:"gpupconfig" env:"GPUPCONFIG" default:"~/.gpupconfig" description:"Path to the config file"`
	NewAlbum   string `short:"n" long:"new-album" value-name:"TITLE" description:"Create an album and add files into it"`
	Debug      bool   `long:"debug" env:"DEBUG" description:"Enable request and response logging"`

	ExternalConfig ExternalConfig `group:"Options read from gpupconfig"`

	Paths []string
}

// New creates a new CLI object.
//
// This does the followings:
// - Determine path to the config
// - Read the config
// - Parse the arguments
// - Validate
//
// If the config is invalid, it will be ignored.
func New(osArgs []string, version string) (*CLI, error) {
	var c CLI
	parser := flags.NewParser(&c, flags.HelpFlag)
	parser.Usage = "[OPTIONS] FILE or DIRECTORY..."
	parser.LongDescription = fmt.Sprintf("Version %s", version)
	if _, err := parser.ParseArgs(osArgs[1:]); err != nil {
		return nil, err
	}
	if err := c.ExternalConfig.Read(c.ConfigName); err != nil {
		log.Printf("Skip reading %s: %s", c.ConfigName, err)
	}
	var err error
	c.Paths, err = parser.ParseArgs(osArgs[1:])
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Run runs the command.
func (c *CLI) Run(ctx context.Context) error {
	if c.ExternalConfig.ClientID == "" || c.ExternalConfig.ClientSecret == "" {
		if err := c.initialSetup(ctx); err != nil {
			return err
		}
	}
	switch {
	case len(c.Paths) == 0:
		return fmt.Errorf("Nothing to upload")
	case c.NewAlbum != "":
		return c.createAlbum(ctx)
	default:
		return c.addToLibrary(ctx)
	}
}

func (c *CLI) initialSetup(ctx context.Context) error {
	log.Printf(`Setup your API access by the following steps:

1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
1. Enable Photos Library API.
1. Open https://console.cloud.google.com/apis/credentials
1. Create an OAuth client ID where the application type is other.

`)
	fmt.Printf("Enter your OAuth client ID (e.g. xxx.apps.googleusercontent.com): ")
	fmt.Scanln(&c.ExternalConfig.ClientID)
	if c.ExternalConfig.ClientID == "" {
		return fmt.Errorf("OAuth client ID must not be empty")
	}
	fmt.Printf("Enter your OAuth client secret: ")
	fmt.Scanln(&c.ExternalConfig.ClientSecret)
	if c.ExternalConfig.ClientSecret == "" {
		return fmt.Errorf("OAuth client ID must not be empty")
	}
	if err := c.ExternalConfig.Write(c.ConfigName); err != nil {
		return fmt.Errorf("Could not save credentials to %s: %s", c.ConfigName, err)
	}
	log.Printf("Saved credentials to %s", c.ConfigName)
	return nil
}
