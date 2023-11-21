// Package field implements various fields (algebraic structures).
// I.e a set with four operations: addition, subtraction, multiplication
// and division.
package field

import (
	"fmt"
	"math/big"
)

// BitLength is the max bitsize for the integer used to define the
// order of the field.
const BitLength = 64

var (
	zero = big.NewInt(0)
	one  = big.NewInt(1)
)

// Finite defines a finite prime field.
// For simplicity it's using a 64bit integer.
// Internal calculations are done using big.Int to avoid any overflows.
type Finite struct {
	p  int64 // Order of the finite field
	zp *big.Int
}

// NewFinite initializes and returns a finite field.
func NewFinite(p int64) *Finite {
	return &Finite{
		p:  p,
		zp: big.NewInt(p),
	}
}

// P returns the order of the finit field.
func (f *Finite) P() int64 {
	return f.p
}

// Element returns true if the provided element is a member of the field.
func (f Finite) Element(i int64) bool {
	if i < 0 {
		return false
	}

	if i >= f.p {
		return false
	}

	return true
}

// Canonicalize returns the canonical representation of an element.
// The canonical representation is always positive and mod P.
func (f Finite) Canonicalize(i int64) int64 {
	i %= f.p

	if i < 0 {
		i += f.p
	}

	return i
}

// canonicalize a *big.Int representation of an element.
func (f Finite) canonicalize(z *big.Int) int64 {
	z.Mod(z, f.zp)

	if z.Cmp(zero) < 0 {
		z.Add(z, f.zp)
	}

	return z.Int64()
}

// Add to elements and return the canonicalize result.
func (f Finite) Add(i, j int64) int64 {
	var z1 = big.NewInt(i)
	var z2 = big.NewInt(j)

	return f.canonicalize(z1.Add(z1, z2))
}

// Multiply two elements and return the result.
func (f Finite) Multiply(i, j int64) int64 {
	var z1 = big.NewInt(i)
	var z2 = big.NewInt(j)

	return f.canonicalize(z1.Mul(z1, z2))
}

// Inverse computes the multiplicative inverse of i mod n
// nolint: lll
// nolint: revive
// See https://en.wikipedia.org/wiki/Extended_Euclidean_algorithm#Computing_multiplicative_inverses_in_modular_structures
// for reference.
func (f Finite) Inverse(i int64) (int64, error) {
	var t = big.NewInt(0)
	var r = big.NewInt(f.p)
	var newT = big.NewInt(1)
	var newR = big.NewInt(i % f.p)

	for newR.BitLen() != 0 {
		var q big.Int
		var tmpT big.Int
		var tmpR big.Int

		q.Div(r, newR)

		tmpT.Mul(&q, newT)
		tmpT.Sub(t, &tmpT)
		t = newT
		newT = &tmpT

		tmpR.Mul(&q, newR)
		tmpR.Sub(r, &tmpR)
		r = newR
		newR = &tmpR
	}

	if r.Cmp(one) > 0 {
		return 0, fmt.Errorf("%d is not invertible", i)
	}

	return f.canonicalize(t), nil
}

// Exponentiate raises i to the power of j mod n.
func (f Finite) Exponentiate(i, j int64) int64 {
	var r int64 = 1

	if i >= f.p {
		panic(i)
	}

	if j >= f.p {
		panic(j)
	}

	for b := 0; b < BitLength; b++ {
		if (j & (1 << b)) != 0 {
			r = f.Multiply(r, i)
		}

		i = f.Multiply(i, i)
	}

	return f.Canonicalize(r)
}

// Sqrt computes the square root of i in this field.
// This operation only works if N is congruent 3 mod 4.
// The sqrt are +/- the returned value.
// If the field is over a number not congruent 3 mod 4, or i is not square
// mod N, an error is returned.
func (f Finite) Sqrt(i int64) (int64, error) {
	var rem = f.p % 4

	if i >= f.p {
		panic(i)
	}

	if rem != 3 {
		return 0, fmt.Errorf("Sqrt not implemented for this field n: %d", f.p)
	}

	if i == 0 {
		return 0, nil
	}

	// This is not trivial. But we know (= should be read as congruent
	// mod N):
	// * P = 3 mod 4
	// * i^(p-1) = 1 (Fermat's little theorem)
	// Now let x = i^((p+1)/4), then:
	// x^4 = i^(p+1) = i^2 * i^(p-1) = i^2
	// This implies (x^2 + i)(x^2 - i) = 0
	// and so the square root of i is +/- x
	var e = (f.p + 1) / 4 // This works as P is congruent 3 mod 4
	var x = f.Exponentiate(i, e)
	// if i has a square root, x^2 = i, verify
	var j = f.Multiply(x, x)

	if i == j {
		return x, nil
	}

	return 0, fmt.Errorf("%d is not a square mod %d", i, f.p)
}
