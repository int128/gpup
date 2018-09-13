package main

import (
	"context"
	"log"
	"os"

	"github.com/int128/gpup/cli"
)

// Set by goreleaser, see https://goreleaser.com/environment/
var version = "1.x"

func main() {
	c, err := cli.New(os.Args, version)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if err := c.Run(ctx); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
