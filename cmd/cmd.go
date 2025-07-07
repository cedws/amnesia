package cmd

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/x/term"
)

type cli struct {
	Seal   sealCmd   `cmd:""`
	Unseal unsealCmd `cmd:""`
}

func haveStdin() bool {
	return !term.IsTerminal(uintptr(os.Stdin.Fd()))
}

func Execute() error {
	var cli cli

	ctx := kong.Parse(
		&cli,
		kong.Name("amnesia"),
		kong.Description("Tool for sealing and unsealing secrets with a set of questions"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	return ctx.Run()
}
