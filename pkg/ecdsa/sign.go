// Package ecdsa implements ECDSA signing.
package ecdsa

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/kommendorkapten/sigsim/pkg/ec"
	"github.com/kommendorkapten/sigsim/pkg/field"
)

// PrivateKey on an elliptic curve.
type PrivateKey struct {
	Pub *PublicKey
	D   int64
}

// PublicKey for a key on an elliptic curve.
type PublicKey struct {
	C *ec.Curve
	P ec.Point
}

var errInvK = errors.New("invalid k for given curve")

func generateKey(c *ec.Curve, d int64) *PrivateKey {
	var pub = PublicKey{
		C: c,
		P: c.ScalarM(d, c.G),
	}
	var p = PrivateKey{
		Pub: &pub,
		D:   d,
	}

	return &p
}

// GenerateKey returns a randomly generated private key on the provided curve.
func GenerateKey(c *ec.Curve, r io.Reader) (*PrivateKey, error) {
	var max = big.NewInt(c.N)
	var zero = big.NewInt(0)

	for {
		var d *big.Int
		var err error

		d, err = rand.Int(r, max)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random number %w", err)
		}

		if d.Cmp(zero) > 0 {
			return generateKey(c, d.Int64()), nil
		}
	}
}

// Sign a message. Returned is the r and s values.
func Sign(r io.Reader, p *PrivateKey, h []byte) (int64, int64, error) {
	var k *big.Int

	for {
		var err error

		k, err = rand.Int(r, big.NewInt(p.Pub.C.N))
		if err != nil {
			return 0, 0, fmt.Errorf("failed generate random number %w", err)
		}

		if k.BitLen() > 0 {
			break
		}
	}

	return rawSign(p, k.Int64(), h)
}

// Calculate a signature given a k and a digest:
// z = h truncated to the curve's bitlength
// p = k*G
// r = p.X mod N
// s = (z + r*d)/k mod N
// If r is congruent to 0 mod N, error is returned
// If k does not invertible mod N, error is returned
// If s is 0, error is returned
// The signature pair (r, s) is returned.
func rawSign(pk *PrivateKey, k int64, h []byte) (int64, int64, error) {
	var r, s, z, inv int64
	// Signatures are computed mod N
	var sf = field.NewFinite(pk.Pub.C.N)
	var p ec.Point
	var err error

	z = truncate(h, pk.Pub.C.BS)
	p = pk.Pub.C.ScalarM(k, pk.Pub.C.G)

	if r = sf.Canonicalize(p.X); r == 0 {
		return 0, 0, errInvK
	}

	// This shouldn't happen if N is prime
	if inv, err = sf.Inverse(k); err != nil {
		return 0, 0, errInvK
	}

	s = sf.Multiply(inv, sf.Add(z, sf.Multiply(r, pk.D)))
	if s == 0 {
		return 0, 0, errInvK
	}

	return r, s, nil
}

// Verify a signature (r and s) for a give message.
func Verify(pub *PublicKey, r, s int64, h []byte) bool {
	var z int64
	var u1, u2, inv int64
	// Signatures are compute mod N
	var sf = field.NewFinite(pub.C.N)
	var cp ec.Point
	var err error

	// Public key must not be the identity element
	if pub.P.Inf {
		return false
	}
	// Point must be on the curve
	if !pub.C.Valid(pub.P) {
		return false
	}
	// order * point must be the identity element
	var q = pub.C.ScalarM(pub.C.N, pub.P)
	if !q.Inf {
		return false
	}

	// Public key is valid, verify the signature
	if r < 1 || r >= pub.C.N {
		return false
	}

	if s < 1 || s >= pub.C.N {
		return false
	}

	z = truncate(h, pub.C.BS)

	inv, err = sf.Inverse(s)
	if err != nil {
		return false
	}

	u1 = sf.Multiply(z, inv)
	u2 = sf.Multiply(r, inv)

	cp = pub.C.Add(pub.C.ScalarM(u1, pub.C.G), pub.C.ScalarM(u2, pub.P))
	if cp.Inf {
		return false
	}

	// Signature is valid if cp.X is congruent to r mod N
	// cp is calculated over the curve of order P which may be higher
	// than the sub-field N
	return sf.Canonicalize(cp.X) == r
}

func solve(c *ec.Curve, r, s1, s2 int64, h1, h2 []byte) (int64, error) {
	// k = (z2 - z1) / (s2 - s1)
	// d = (s1k - z1) / r
	var z1 = truncate(h1, c.BS)
	var z2 = truncate(h2, c.BS)
	var sf = field.NewFinite(c.N)

	var inv, err = sf.Inverse(sf.Add(s2, -s1))
	if err != nil {
		return 0, fmt.Errorf("could not inverese s2 - s1: %w", err)
	}

	var k = sf.Multiply(sf.Add(z2, -z1), inv)

	inv, err = sf.Inverse(r)
	if err != nil {
		return 0, fmt.Errorf("could not inverse r: %w", err)
	}

	var d1 = sf.Multiply(sf.Add(sf.Multiply(s1, k), -z1), inv)
	var d2 = sf.Multiply(sf.Add(sf.Multiply(s2, k), -z2), inv)

	if d1 != d2 {
		return 0, fmt.Errorf("could not compute d: %d != %d",
			d1, d2)
	}

	return d1, nil
}

// Truncate treats b as a big endian integer.
// Returns the bs most significant bits.
func truncate(b []byte, bs int) int64 {
	var z = new(big.Int)
	var nb = (bs + 7) / 8
	var rem = uint((8 - (bs % 8)) % 8)

	if bs > 31 {
		panic(bs)
	}

	z.SetBytes(b[:nb])

	if rem > 0 {
		// Rotate right to get rid of the extra bytes
		z.Rsh(z, rem)
	}

	return z.Int64()
}
