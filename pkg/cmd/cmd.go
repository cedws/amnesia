package cmd

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/cedws/amnesia/pkg/amnesia/ageplugin"
	"github.com/charmbracelet/x/term"
)

type cli struct {
	Seal      sealCmd      `cmd:""`
	Unseal    unsealCmd    `cmd:""`
	AgeKeygen ageKeygenCmd `cmd:""`
}

func haveStdin() bool {
	return !term.IsTerminal(uintptr(os.Stdin.Fd()))
}

func Execute() error {
	if os.Args[0] == "age-plugin-amnesia" {
		os.Exit(ageplugin.Main())
	}

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
