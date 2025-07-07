package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/cedws/amnesia/amnesia"
	"github.com/cedws/amnesia/amnesia/interactive"
)

type sealCmd struct {
	File       string `help:"File to write sealed secret to." short:"f"`
	NoCompress bool   `help:"Don't compress the secret with gzip before sealing." short:"c"`
	NoTest     bool   `help:"Don't prompt for test questions." short:"t"`
}

func (s *sealCmd) Help() string {
	return `Seal a secret passed via stdin.

This command reads sensitive data from stdin and encrypts it using a set of questions and answers. The secret is split using Shamir's Secret Sharing algorithm, where each question/answer pair protects one share.

Examples:
  echo "my secret password" | amnesia seal
  cat ~/.ssh/id_rsa | amnesia seal -f sealed.json
  amnesia seal --no-compress -f sealed.json < large-file.txt`
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
	if !s.NoCompress {
		opts = append(opts, interactive.WithSealOptions(
			amnesia.WithCompression(),
		))
	}

	sealed, err := interactive.Seal(context.Background(), data, opts...)
	if err != nil {
		return err
	}

	if s.File != "" {
		if err := os.WriteFile(s.File, sealed, 0644); err != nil {
			return err
		}

		return nil
	}

	if _, err := os.Stdout.Write(sealed); err != nil {
		return err
	}

	return nil
}
