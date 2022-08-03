package creds

import (
	"os"

	"github.com/nats-io/nkeys"
	"github.com/pkg/errors"
)

func Parse(data []byte) (token string, public string, seed []byte, err error) {
	token, err = nkeys.ParseDecoratedJWT(data)
	if err != nil {
		return "", "", []byte{}, errors.Wrap(err, "failed to parse jwt")
	}
	nkey, err := nkeys.ParseDecoratedNKey(data)
	if err != nil {
		return "", "", []byte{}, errors.Wrap(err, "failed to parse nkey")
	}

	// this turns the keypair seed into junk data, so not sure we can wipe here
	//defer nkey.Wipe()

	public, err = nkey.PublicKey()
	if err != nil {
		return "", "", []byte{}, errors.Wrap(err, "failed to get nkey public key")
	}

	seed, err = nkey.Seed()
	if err != nil {
		return "", "", []byte{}, errors.Wrap(err, "failed to get nkey seed")
	}

	return token, public, seed, nil
}

func ParseFile(f string) (token string, public string, seed []byte, err error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return "", "", []byte{}, errors.Wrapf(err, "failed to read file %q", f)
	}

	return Parse(data)
}
