package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"playground/cmd/creds"

	jwt "github.com/nats-io/jwt/v2"
	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/pkg/errors"
)

var (
	address        = flag.String("address", nats.DefaultURL, "NATS address")
	credentialFile = flag.String("creds", "./nsc/keys/creds/local/SYS/sys.creds", "Credentials")
	credentialDir  = flag.String("dir", "./nsc/store", "Where the nsc store is located")

	// ghetto claims db
	claims = sync.Map{}
)

const (
	// haven't gotten this to trigger yet, i think because all nats servers are `full` resolvers, this won't get hit
	// will have to update to a few of them being cache instead of full
	CLAIMS_LOOKUP = `$SYS.REQ.ACCOUNT.*.CLAIMS.LOOKUP`

	// nsc push will trigger this
	CLAIMS_UPDATE = `$SYS.REQ.CLAIMS.UPDATE`
)

func main() {
	flag.Parse()

	if err := loadClaims(*credentialDir); err != nil {
		printE(err)
	}

	token, nkeyPublic, nkeySeed, err := creds.ParseFile(*credentialFile)
	if err != nil {
		printE(err)
	}

	print("Public Key: %s", nkeyPublic)
	print("Private Key: %s", string(nkeySeed))

	opts := nats.Options{
		AllowReconnect:     true,
		MaxReconnect:       -1, // connect forever
		ReconnectWait:      nats.DefaultReconnectWait,
		ReconnectJitter:    nats.DefaultReconnectJitter,
		ReconnectJitterTLS: nats.DefaultReconnectJitterTLS,
		Timeout:            nats.DefaultTimeout,
		PingInterval:       nats.DefaultPingInterval,
		MaxPingsOut:        nats.DefaultMaxPingOut,
		SubChanLen:         nats.DefaultMaxChanLen,
		ReconnectBufSize:   nats.DefaultReconnectBufSize,
		DrainTimeout:       nats.DefaultDrainTimeout,
		Servers:            []string{*address},
		Name:               "Demo Client",
		AsyncErrorCB: func(conn *nats.Conn, sub *nats.Subscription, err error) {
			print("Slow subscriber %q", sub.Subject)
		},
		ClosedCB: func(conn *nats.Conn) {
			print("Connection to server %q closed", conn.ConnectedServerName())
		},
		DisconnectedErrCB: func(conn *nats.Conn, err error) {
			if err != nil {
				print("Disconnected from server %q: %v", conn.ConnectedServerName(), err)
			}
		},
		DiscoveredServersCB: func(conn *nats.Conn) {
			print("Discovered new server: %q", conn.ConnectedServerName())
		},
		UserJWT: func() (string, error) {
			return token, nil
		},
		ReconnectedCB: func(conn *nats.Conn) {
			print("Reconnected to server %q", conn.ConnectedServerName())
		},
		SignatureCB: func(data []byte) ([]byte, error) {
			key, err := nkeys.FromSeed(nkeySeed)
			if err != nil {
				printE(err)
			}

			return key.Sign(data)
		},
	}

	// super derpy reconnect when servers aren't up yet
RETRY_CONNECT:
	nc, err := opts.Connect()
	if errors.Is(err, nats.ErrNoServers) {
		time.Sleep(time.Millisecond * 100)
		goto RETRY_CONNECT
	}

	if err != nil {
		printE(err)
	}
	defer func() { _ = nc.Drain() }()
	defer nc.Close()

	print("%v to Server %s", nc.Status(), nc.ConnectedServerName())

	// subscribe to claims requests
	claimsRequestSub, err := nc.Subscribe(CLAIMS_LOOKUP, func(msg *nats.Msg) {
		print("CLAIMS LOOKUP: %s", msg.Subject)
		subjectParts := strings.Split(msg.Subject, ".")
		if len(subjectParts) != 6 {
			print("Subject %q did not match intended claim lookup", msg.Subject)
			return
		}

		accountID := subjectParts[3]
		// we have a valid account id, just need to respond with the JWT (encoded, not json)
		print("AccountID: %q; Responder: %q; Headers: %+v", accountID, msg.Reply, msg.Header)

		claim, found := claims.Load(accountID)
		if !found {
			print("Claim not found: %s", accountID)
			return
		}

		if err := msg.Respond(claim.([]byte)); err != nil {
			print("Failed to respond to claim request: %v", err)
		}
	})

	if err != nil {
		printE(err)
	}
	defer func() { _ = claimsRequestSub.Drain() }()
	defer func() { _ = claimsRequestSub.Unsubscribe() }()

	// subscribe to claims updates
	claimsUpdateSub, err := nc.Subscribe(CLAIMS_UPDATE, func(msg *nats.Msg) {
		print("CLAIMS UPDATE: %s\n%s", msg.Subject, string(msg.Data))
		c, err := jwt.DecodeAccountClaims(string(msg.Data))
		if err != nil {
			print("CLAIMS FAILURE: Could not parse JWT: %v", err)
			return
		}

		// TODO: should probably validate the jwt...
		print("JWT: %s", c)

		// store the claim by account id so it can be retrieved by the claims lookup
		claims.Store(c.Subject, msg.Data)
		if err := msg.Respond([]byte("jwt updated")); err != nil {
			print("CLAIM UPDATE ERROR: %v", err)
		}
	})

	if err != nil {
		printE(err)
	}
	defer func() { _ = claimsUpdateSub.Drain() }()
	defer func() { _ = claimsUpdateSub.Unsubscribe() }()

	schan := make(chan os.Signal, 6)
	defer close(schan)

	signal.Notify(schan, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGHUP, syscall.Signal(21))
	defer signal.Stop(schan)

	<-schan
}

func loadClaims(dir string) error {
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".jwt" {
			return nil
		}

		jwtData, readErr := os.ReadFile(path)
		if readErr != nil {
			return errors.Wrapf(readErr, "failed to read jwt %q", path)
		}

		claim, jwtErr := jwt.DecodeAccountClaims(string(jwtData))
		if jwtErr != nil {
			if jwtErr.Error() == "not account claim" {
				return nil
			}
			return errors.Wrapf(jwtErr, "failed to decode account claims %q", path)
		}

		claims.Store(claim.Subject, jwtData)

		print("Claim Loaded: %s from file %s", claim.Subject, path)

		return nil
	})
}

func print(m string, args ...any) {
	now := time.Now()
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	microsec := now.Nanosecond() / 1000
	ts := fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d.%06d",
		year, month, day, hour, min, sec, microsec)

	fmt.Fprintf(os.Stdout, ts+" "+m+"\n", args...)
}

func printE(err error) {
	fmt.Fprintf(os.Stderr, "Error %T: %s\n", err, err.Error())
	os.Exit(1)
}
