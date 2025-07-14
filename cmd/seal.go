package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/cedws/amnesia/amnesia/interactive"
)

type sealCmd struct {
	File   string `help:"File to write sealed secret to." short:"f"`
	NoTest bool   `help:"Don't prompt for test questions." short:"t"`
}

func (s *sealCmd) Help() string {
	return `Seal a secret passed via stdin.

This command reads sensitive data from stdin and encrypts it using a set of questions and answers. The secret is split using Shamir's Secret Sharing algorithm, where each question/answer pair protects one share.

Examples:
  echo "my secret password" | amnesia seal
  cat ~/.ssh/id_rsa | amnesia seal -f sealed.json
  amnesia seal -f sealed.json < large-file.txt`
}

func (s *sealCmd) AfterApply() error {
	if !haveStdin() {
		return fmt.Errorf("no data passed to stdin")
	}

	return nil
}

func (s *sealCmd) Run(ctx *kong.Context) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	var opts []interactive.Option

	if !s.NoTest {
		opts = append(opts, interactive.WithTestQuestions())
	}

	sealed, err := interactive.Seal(context.Background(), data, opts...)
	if err != nil {
		return fmt.Errorf("failed to seal secret: %w", err)
	}

	if s.File != "" {
		if err := os.WriteFile(s.File, sealed, 0600); err != nil {
			return err
		}

		return nil
	}

	if _, err := os.Stdout.Write(sealed); err != nil {
		return err
	}

	return nil
}
