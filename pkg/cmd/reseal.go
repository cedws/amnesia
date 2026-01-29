package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/cedws/amnesia/pkg/amnesia/interactive"
	"github.com/charmbracelet/x/term"
)

type resealCmd struct {
	File       string `help:"File to reseal secret from." short:"f"`
	OutputFile string `help:"File to write resealed secret to." short:"o"`
}

func (r *resealCmd) Help() string {
	return `Reseal a new secret using an existing sealed file's questions.

This command reads a new secret from stdin and encrypts it using the questions from an existing sealed file. You must provide the correct answers to derive the encryption key.

Examples:
  echo "new secret" | amnesia reseal -f sealed.json
  echo "new secret" | amnesia reseal -f sealed.json -o resealed.json
  cat new-secret.txt | amnesia reseal -f sealed.json`
}

func (r *resealCmd) AfterApply() error {
	if !haveStdin() {
		return fmt.Errorf("no data passed to stdin")
	}
	if term.IsTerminal(uintptr(os.Stdin.Fd())) && r.File == "" {
		return fmt.Errorf("file is required when not reading from stdin")
	}

	return nil
}

func (r *resealCmd) Run(ctx *kong.Context) error {
	sealed, err := os.ReadFile(r.File)
	if err != nil {
		return err
	}

	newSecret, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	resealed, err := interactive.Reseal(context.Background(), sealed, newSecret)
	if err != nil {
		return fmt.Errorf("failed to reseal secret: %w", err)
	}

	if r.OutputFile != "" {
		if err := os.WriteFile(r.OutputFile, resealed, 0600); err != nil {
			return err
		}

		return nil
	}

	if _, err := os.Stdout.Write(resealed); err != nil {
		return err
	}

	return nil
}
