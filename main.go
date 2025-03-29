package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Chat Chat `kong:"cmd,help='Start an interactive chat with AI'"`
	Auth Auth `kong:"cmd,help='Authenticate your Devgraph client'"`
}

func main() {

	cli := CLI{}

	ctx := kong.Parse(&cli,
		kong.Name("devgraph"),
		kong.Description("Turn chaos into clarity"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
