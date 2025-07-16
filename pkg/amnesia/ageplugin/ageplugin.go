package ageplugin

import (
	"context"
	"fmt"

	"filippo.io/age"
	"filippo.io/age/plugin"
	"github.com/cedws/amnesia/pkg/amnesia"
	"github.com/cedws/amnesia/pkg/amnesia/interactive"
)

const pluginName = "amnesia"

func Main() int {
	plugin, err := plugin.New(pluginName)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	plugin.HandleIdentity(func(data []byte) (age.Identity, error) {
		return identityPlugin{
			data:   data,
			plugin: plugin,
		}, nil
	})

	return plugin.Main()
}

type identityPlugin struct {
	plugin *plugin.Plugin
	data   []byte
}

func (i identityPlugin) unwrap(stanzas []*age.Stanza) ([]byte, error) {
	sealedSecret, err := amnesia.Decode(i.data)
	if err != nil {
		return nil, err
	}

	answers := amnesia.NewAnswers()

	for _, share := range sealedSecret.Shares {
		s := fmt.Sprintf("amnesia: Enter answer to question\n%s:", share.Question)
		answer, err := i.plugin.RequestValue(s, false)
		if err != nil {
			return nil, err
		}

		answers.Set(share.ID, answer)
	}

	unsealed, err := amnesia.Unseal(i.data, answers)
	if err != nil {
		return nil, fmt.Errorf("error unsealing key (incorrect or too few answers?)")
	}

	identity, err := age.ParseX25519Identity(string(unsealed))
	if err != nil {
		return nil, err
	}

	return identity.Unwrap(stanzas)
}

func (i identityPlugin) Unwrap(stanzas []*age.Stanza) ([]byte, error) {
	identityKey, err := i.unwrap(stanzas)
	if err != nil {
		i.plugin.DisplayMessage(err.Error())
		return nil, age.ErrIncorrectIdentity
	}

	return identityKey, nil
}

type Identity struct {
	ageIdentity  *age.X25519Identity
	sealedSecret []byte
}

func (i Identity) Identity() string {
	return plugin.EncodeIdentity(pluginName, i.sealedSecret)
}

func (i Identity) Recipient() string {
	return i.ageIdentity.Recipient().String()
}

func GenerateIdentity(ctx context.Context, opts ...interactive.Option) (Identity, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return Identity{}, err
	}

	sealed, err := interactive.Seal(ctx, []byte(identity.String()), opts...)
	if err != nil {
		return Identity{}, err
	}

	return Identity{
		sealedSecret: sealed,
		ageIdentity:  identity,
	}, nil
}
