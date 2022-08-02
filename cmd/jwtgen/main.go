package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/pkg/errors"
)

var (
	nkeyFile = flag.String("nkey", "./nkey.APP", "Nkey file")
)

func main() {
	flag.Parse()

	nkeyBits, err := os.ReadFile(*nkeyFile)
	if err != nil {
		printE(errors.Wrapf(err, "failed to read nkey file %s", *nkeyFile))
	}

	nkey, err := nkeys.ParseDecoratedNKey(nkeyBits)
	if err != nil {
		printE(errors.Wrap(err, "failed to parse nkey"))
	}

	nkeyPublic, err := nkey.PublicKey()
	if err != nil {
		printE(errors.Wrap(err, "failed to read nkey public key"))
	}
	nkeySeed, err := nkey.Seed()
	if err != nil {
		printE(errors.Wrap(err, "failed to read nkey seed"))
	}

	userClaims := jwt.NewUserClaims(nkeyPublic)
	userClaims.Name = "leaf-node"
	userClaims.IssuerAccount = "APP"

	userJwt, err := userClaims.Encode(nkey)
	if err != nil {
		printE(errors.Wrap(err, "unable to encode claim"))
	}

	userConfig, err := jwt.FormatUserConfig(userJwt, nkeySeed)
	if err != nil {
		printE(errors.Wrap(err, "failed to format user config"))
	}

	print(string(userConfig))

}

func print(m string, args ...any) {
	fmt.Fprintf(os.Stdout, m+"\n", args...)
}

func printE(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(1)
}
