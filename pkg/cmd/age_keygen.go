package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/cedws/amnesia/pkg/amnesia/ageplugin"
	"github.com/cedws/amnesia/pkg/amnesia/interactive"
)

type ageKeygenCmd struct {
	OutputFile string `help:"File to write the identity to." short:"o" name:"output"`
	NoTest     bool   `help:"Don't prompt for test questions." short:"t"`
}

func (s *ageKeygenCmd) Help() string {
	return `Generate an amnesia-compatible age identity (experimental)

This command generates an amnesia-sealed age X25519 identity which can be used with age when amnesia is installed as a plugin.

Examples:
  amnesia age-keygen
  amnesia age-keygen -o identity.txt`
}

func (s *ageKeygenCmd) interactiveOpts() []interactive.Option {
	var opts []interactive.Option

	if !s.NoTest {
		opts = append(opts, interactive.WithTestQuestions())
	}

	return opts
}

func (s *ageKeygenCmd) Run(ctx *kong.Context) error {
	identity, err := ageplugin.GenerateIdentity(context.Background(), s.interactiveOpts()...)
	if err != nil {
		return err
	}

	if s.OutputFile != "" {
		if err := os.WriteFile(s.OutputFile, []byte(identity.Identity()), 0600); err != nil {
			return err
		}

		fmt.Printf("Public key: %s\n", identity.Recipient())

		return nil
	}

	fmt.Printf("# created %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("# public key: %s\n", identity.Recipient())
	fmt.Println(identity.Identity())

	return nil
}
