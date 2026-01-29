package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/alecthomas/kong"
	"github.com/cedws/amnesia/pkg/amnesia"
	"github.com/cedws/amnesia/pkg/amnesia/interactive"
	"github.com/gofrs/flock"
)

type openCmd struct {
	File       string `help:"File to read sealed secret from." required:"true" short:"f"`
	SecretFile string `help:"File to write secret to and reseal later."  required:"true" short:"o"`
}

func (o *openCmd) Help() string {
	return ``
}

func (o *openCmd) Run(ctx *kong.Context) error {
	lock, err := o.lockSecret()
	if err != nil {
		return err
	}
	defer lock.Close()

	sealed, err := os.ReadFile(o.File)
	if err != nil {
		return err
	}

	key, err := interactive.DecryptKey(context.Background(), sealed)
	if err != nil {
		return err
	}

	secret, err := amnesia.UnsealWithKey(sealed, key)
	if err != nil {
		return err
	}

	if err := os.WriteFile(o.SecretFile, secret, 0600); err != nil {
		return err
	}
	defer os.Remove(o.SecretFile)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	fmt.Println("Awaiting ^C to reseal...")
	<-signalCh

	if err := o.reseal(sealed, key); err != nil {
		return err
	}

	return nil
}

func (o *openCmd) lockSecret() (io.Closer, error) {
	sealedLock := flock.New(o.File)
	locked, err := sealedLock.TryLock()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, fmt.Errorf("file %s is locked by another process", o.File)
	}

	return sealedLock, nil
}

func (o *openCmd) reseal(sealed, key []byte) error {
	secret, err := os.ReadFile(o.SecretFile)
	if err != nil {
		return err
	}

	newSealed, err := amnesia.ResealWithKey(sealed, secret, key)
	if err != nil {
		return err
	}

	if err := os.WriteFile(o.File, newSealed, 0600); err != nil {
		return err
	}

	return nil
}
