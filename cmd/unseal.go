package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/cedws/amnesia/amnesia/interactive"
	"github.com/charmbracelet/x/term"
)

type unsealCmd struct {
	File       string `help:"File to unseal secret from." short:"f"`
	OutputFile string `help:"File to write unsealed secret to." short:"o"`
}

func (s *unsealCmd) Help() string {
	return `Unseal a secret.

This command reconstructs a secret by prompting for answers to the questions that were set during sealing. You must provide the minimum threshold number of correct answers to successfully unseal the secret. The sealed data can be provided via stdin or from a file.

Examples:
  amnesia unseal < sealed.json
  amnesia unseal -f sealed.json
  amnesia unseal -f sealed.json -o recovered-secret.txt
  cat sealed.json | amnesia unseal -o original-file.txt`
}

func (s *unsealCmd) AfterApply() error {
	if term.IsTerminal(uintptr(os.Stdin.Fd())) && s.File == "" {
		return fmt.Errorf("file is required when not reading from stdin")
	}

	return nil
}

func (s *unsealCmd) Run(ctx *kong.Context) error {
	input, err := readInput(s)
	if err != nil {
		return err
	}

	unsealed, err := interactive.Unseal(context.Background(), input)
	if err != nil {
		return err
	}

	if s.OutputFile != "" {
		if err = os.WriteFile(s.OutputFile, unsealed, 0644); err != nil {
			return err
		}

		return nil
	}

	if _, err := os.Stdout.Write(unsealed); err != nil {
		return err
	}

	return nil
}

func readInput(cmd *unsealCmd) ([]byte, error) {
	if !haveStdin() {
		return os.ReadFile(cmd.File)
	}

	buf, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
