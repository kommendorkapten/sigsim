// Package app contains the implementation of the various commands.
package app

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"

	"github.com/kommendorkapten/sigsim/pkg/ec"
	"github.com/kommendorkapten/sigsim/pkg/field"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// candidate curve
// 6978607847 6978607844 171221
// order 6978671083

// GenCurve generates a random curve based on the provieded parameters.
// The default parameter for a should normally not be changed.
func GenCurve() *ffcli.Command {
	var (
		flagset = flag.NewFlagSet("sigsim genc", flag.ExitOnError)
		p       = flagset.Int("p", 33489583, "Order of the finte field")
		a       = flagset.Int("a", -3, "A parameter for curve")
		b       = flagset.Int("b", 3411011, "B parameter for curve")
		o       = flagset.Int("o", 0, "Offset to start search in")
	)

	return &ffcli.Command{
		Name:       "genc",
		ShortUsage: "sigsim genc",
		ShortHelp:  "Generate a curve",
		LongHelp:   "Generate a curve",
		FlagSet:    flagset,
		Exec: func(ctx context.Context, args []string) error {
			return GenCurveCmd(ctx,
				int64(*p),
				int64(*a),
				int64(*b),
				int64(*o),
			)
		},
	}
}

func GenCurveCmd(_ context.Context, n, a, b, o int64) error {
	// we want to have a curve that satisfies:
	// 1. order is a prime
	// 2. Generator point that gives a cofactor of 1
	// 3. Order of generator point to be a prime

	if a < 0 {
		a += n
	}

	if b < 0 {
		b += n
	}

	var c *ec.Curve
	var err error
	for {
		c, err = ec.NewCurve(field.NewFinite(n), a, b)
		if err != nil {
			return err
		}

		fmt.Printf("Test curve %+v\n", c)

		// This call takes a long time!
		var order = c.CountPoints()
		fmt.Printf("curve has order %d\n", order)

		var p = big.NewInt(order)
		var prime = p.ProbablyPrime(256)
		if prime {
			fmt.Println("Curve order IS prime")
			break
		} else {
			fmt.Println("Curve order is NOT prime")
		}
		b += 2
	}




	// var point = c.RandomPoint()
	// var sgOrder = c.OrderBG(point)
	// fmt.Printf("Order of sug-group %d\n", sgOrder)
	// p = big.NewInt(sgOrder)
	// prime = p.ProbablyPrime(256)
	// if prime {
	// 	fmt.Println("Subgroup order IS prime")
	// } else {
	// 	fmt.Println("Subgroup order is NOT prime")
	// }


	return nil
}

// GenCurveCmd generates the curve.
// Using the curve's parameter is searches for a good generator point and
// halts once one is found.
func GenCurveCmd2(_ context.Context, n, a, b int64) error {
	var c *ec.Curve
	var g ec.Point
	var err error
	var p = 8

	// B should be b^2 * c = -27 mod p
	// c used from a seed (i.e the secret one in NIST p-256.
	if a < 0 {
		a += n
	}

	if b < 0 {
		b += n
	}

	SafePrintf("Searching for generator for curve %d %d %d\n",
		n, a, b)

	// Make sure curve is congruent 3 mod 4.
	if r := n % 4; r != 3 {
		return fmt.Errorf("%d is not congruent 3 mod 4, use another field", n)
	}

	if c, err = ec.NewCurve(field.NewFinite(n), a, b); err != nil {
		return fmt.Errorf("failed to generate curve: %w", err)
	}

	// Estimate a good generator.
	var max = big.NewInt(n)
	var minOrder = int64(float64(n) * 0.45)
	var chn = make(chan ec.Point)

	for i := 0; i < p; i++ {
		go func() {
			var g ec.Point
			var order int64

			for {
				var x *big.Int
				var y int64
				var err error

				x, err = rand.Int(rand.Reader, max)
				if err != nil {
					panic(err)
				}

				y, err = c.Y(x.Int64())
				if err != nil {
					// no valid point
					continue
				}

				g = ec.Point{
					X: x.Int64(),
					Y: y,
				}
				if !c.Valid(g) {
					continue
				}

				order = c.Order(g)
				// SafePrintf("found %d vs %d\n",
				// 	order, n)
				var p = big.NewInt(order)
				var prime = p.ProbablyPrime(256)

				if order > minOrder && prime {
					chn <- g

					break
				}
			}
		}()
	}

	g = <-chn

	SafePrintf("Curve: n: %d a: %d b: %d\n", n, a, b)
	SafePrintf("Generator %+v with order: %d\n", g, c.Order(g))
	for _, i := range c.Points() {
		SafePrintf("%+v\n", i)
	}

	return nil
}
