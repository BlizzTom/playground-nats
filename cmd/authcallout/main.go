package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

func main() {
	emptySeed := [32]byte{}
	seed, err := nkeys.EncodeSeed(nkeys.PrefixByteAccount, emptySeed[:])
	if err != nil {
		panic(err)
	}

	//	accountKP, err := nkeys.CreateAccount()
	accountKP, err := nkeys.FromSeed(seed)
	if err != nil {
		panic(err)
	}

	accountPK, err := accountKP.PublicKey()
	if err != nil {
		panic(err)
	}
	fmt.Println(accountPK)
	fmt.Println(string(seed))
	userKP, err := nkeys.CreateUser()
	if err != nil {
		panic(err)
	}

	userSeed, err := userKP.Seed()
	if err != nil {
		panic(err)
	}

	userPK, err := userKP.PublicKey()
	if err != nil {
		panic(err)
	}

	claims := jwt.NewUserClaims(userPK)
	claims.Name = "Tester"
	claims.Expires = time.Now().Add(time.Hour).Unix()
	claims.IssuedAt = time.Now().Unix()
	claims.Audience = "NATS"

	userToken, err := claims.Encode(accountKP)
	if err != nil {
		panic(err)
	}

	opts := &nats.Options{
		User:     "username",
		Password: "password",
		Token:    "ThisIsMyAuthTokenLolz",
		Url:      "nats://localhost:4222",
	}

	if err := nats.UserJWTAndSeed(userToken, string(userSeed))(opts); err != nil {
		panic(err)
	}

	nc, err := opts.Connect()
	if err != nil {
		panic(err)
	}

	defer nc.Close()

	if err := nc.Publish("foo", []byte("Hello World!")); err != nil {
		slog.Info("Error publishing message", "err", err)
	}

}
