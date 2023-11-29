// Simple command for working with elliptic curves.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kommendorkapten/sigsim/cmd/sigsim/app"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	rootFlagSet = flag.NewFlagSet("sigsim", flag.ExitOnError)
)

func main() {
	root := &ffcli.Command{
		ShortUsage: "sigsim [flags] <subcommand>",
		FlagSet:    rootFlagSet,
		Subcommands: []*ffcli.Command{
			app.GenCurve(),
			app.GenPrime(),
			//			app.Sign(),
		},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.Parse(os.Args[1:]); err != nil {
		printErrAndExit(err)
	}

	if err := root.Run(context.Background()); err != nil {
		printErrAndExit(err)
	}
}

func printErrAndExit(err error) {
	_, err = fmt.Fprintf(os.Stderr, "error: %v\n", err)
	if err != nil {
		panic(err)
	}
	// nolint: revive
	os.Exit(1)
}
