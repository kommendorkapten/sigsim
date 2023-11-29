package app

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// GenPrime returns a command to be used.
func GenPrime() *ffcli.Command {
	var (
		flagset = flag.NewFlagSet("sigsim genp", flag.ExitOnError)
		b       = flagset.Int("b", 25, "Number of bits (max 32)")
		v       = flagset.Bool("v", false, "Verify number is prime")
	)

	return &ffcli.Command{
		Name:       "genp",
		ShortUsage: "sigsim genp",
		ShortHelp:  "Generate a random prime congruent to 3 mod 4",
		LongHelp:   "Generate a random prime contruent to 3 mod 4",
		FlagSet:    flagset,
		Exec: func(ctx context.Context, args []string) error {
			return GenPrimeCmd(ctx, *b, *v)
		},
	}
}

// GenPrimeCmd Generates a prime number of the provided bit length.
// nolint: revive
func GenPrimeCmd(_ context.Context, n int, v bool) error {
	var p *big.Int
	var err error

	for {
		if p, err = rand.Prime(rand.Reader, n); err != nil {
			return fmt.Errorf("failed to acquite a random prime: %w", err)
		}

		if v {
			var start = time.Now()
			var prime = p.ProbablyPrime(64)
			var dur = time.Since(start)

			SafePrintf("Verified prime in %s\n", dur)

			if !prime {
				return fmt.Errorf("generated number was not prime, try again")
			}
		}

		var r big.Int
		r.Mod(p, big.NewInt(4))

		if r.Int64() == 3 {
			break
		}
	}

	SafePrintf("Generated prime: %s\n", p.String())

	return nil
}
