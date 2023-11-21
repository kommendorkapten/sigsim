package ecdsa

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/kommendorkapten/sigsim/pkg/ec"
	"github.com/kommendorkapten/sigsim/pkg/field"
	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	var tests = []struct {
		m   []byte
		bl  int
		exp int
	}{
		{
			m:   []byte{0xff, 0},
			bl:  16,
			exp: 16,
		},
		{
			m:   []byte{0xff, 0},
			bl:  14,
			exp: 14,
		},
		{
			m:   []byte{0xff, 0},
			bl:  15,
			exp: 15,
		},
		{
			m:   []byte{0xff, 0},
			bl:  8,
			exp: 8,
		},
		{
			m:   []byte{0xaa, 0xaa, 0xaa, 0xaa},
			bl:  8,
			exp: 8,
		},
		{
			m:   []byte{0x80, 0x55, 0xaa, 0xaa},
			bl:  24,
			exp: 24,
		},
		{
			m:   []byte{0x55, 0x55, 0xaa, 0xaa},
			bl:  8,
			exp: 7,
		},
		{
			m:   []byte{0x55, 0x55, 0xaa, 0xaa},
			bl:  24,
			exp: 23,
		},
		{
			m:   []byte{0x55, 0x55, 0xaa, 0xaa},
			bl:  13,
			exp: 12,
		},
		{
			m:   []byte{0x85, 0x55, 0xaa, 0xaa},
			bl:  13,
			exp: 13,
		},
		{
			m:   []byte{0x55, 0x55, 0xaa, 0xaa},
			bl:  14,
			exp: 13,
		},
	}

	for _, tc := range tests {
		var z = truncate(tc.m, tc.bl)
		var zb = big.NewInt(z)

		assert.Equal(t, tc.exp, zb.BitLen())
	}

	var h = []byte{0xab, 0xcd, 0xef}

	assert.Equal(t, 0x2af, int(truncate(h, 10)))
}

func TestSignRaw(t *testing.T) {
	t.Run("curve_479_-3_307", func(t *testing.T) {
		var p = generateKey(
			&ec.Curve{
				F: field.NewFinite(479),
				A: -3,
				B: 307,
				G: ec.Point{
					X: 403,
					Y: 280,
				},
				N:  233,
				BS: 8,
			},
			23,
		)
		assert.Nil(t, p.Pub.C.Verify())

		var k int64 = 123
		var h = sha256.Sum256([]byte("a message"))
		r, s, err := rawSign(p, k, h[:])

		assert.Nil(t, err, "got error")
		assert.Equal(t, int64(62), r, "r")
		assert.Equal(t, int64(42), s, "s")

		// sign another message and verify r is the same
		h = sha256.Sum256([]byte("another message"))
		r1, s1, err := rawSign(p, k, h[:])
		assert.Nil(t, err)
		assert.Equal(t, r, r1)
		assert.NotEqual(t, s, s1)

		// Use a k that gives  p.X = 0 (i.e the order of the
		// generator point (actually not a valid k per ECDSA)
		k = 233
		r1, s1, err = rawSign(p, k, h[:])
		assert.Zero(t, r1)
		assert.Zero(t, s1)
		assert.Equal(t, errInvK, err)

		// Now sign again
		k = 123
		r1, s1, err = rawSign(p, k, h[:])
		assert.Nil(t, err)

		// r and s are from the signing of "a message"
		assert.False(t, Verify(p.Pub, r, s, h[:]))
		assert.True(t, Verify(p.Pub, r1, s1, h[:]))
		// Verify first signature
		h = sha256.Sum256([]byte("a message"))
		assert.True(t, Verify(p.Pub, r, s, h[:]))
	})

	t.Run("curve_33489583_-3_3411011", func(t *testing.T) {
		var p = generateKey(ec.DemoCurve25, 847079)
		var h = sha256.Sum256([]byte("important message"))
		var k int64 = 440432
		r, s, err := rawSign(p, k, h[:])

		assert.Nil(t, err)
		assert.True(t, Verify(p.Pub, r, s, h[:]))
	})
}

func TestSign(t *testing.T) {
	t.Run("curve_33489583_-3_3411011", func(t *testing.T) {
		var p = generateKey(ec.DemoCurve25, 847079)

		var h = sha256.Sum256([]byte("important message"))
		r1, s1, err := Sign(rand.Reader, p, h[:])
		assert.Nil(t, err)
		r2, s2, err := Sign(rand.Reader, p, h[:])
		assert.Nil(t, err)

		assert.True(t, Verify(p.Pub, r1, s1, h[:]))
		assert.True(t, Verify(p.Pub, r2, s2, h[:]))

		assert.NotEqual(t, r1, r2)
		assert.NotEqual(t, s1, s2)
	})
}

func TestSolve(t *testing.T) {
	var tests = []PrivateKey{
		{
			Pub: &PublicKey{
				C: &ec.Curve{
					F: field.NewFinite(479),
					A: -3,
					B: 307,
					G: ec.Point{
						X: 403,
						Y: 280,
					},
					N:  233,
					BS: 8,
				},
			},
			D: 23,
		},
		{
			Pub: &PublicKey{
				C: ec.DemoCurve25,
			},
			D: 847079,
		},
	}

	for _, tc := range tests {
		var k int64 = 220
		var h1 = sha256.Sum256([]byte("a message"))
		var h2 = sha256.Sum256([]byte("another message"))

		//nolint: gosec
		r1, s1, err := rawSign(&tc, k, h1[:])

		assert.Nil(t, err)

		//nolint: gosec
		r2, s2, err := rawSign(&tc, k, h2[:])

		assert.Nil(t, err)
		assert.Equal(t, r1, r2)

		d, err := solve(tc.Pub.C, r1, s1, s2, h1[:], h2[:])

		assert.Nil(t, err)

		assert.Equal(t, tc.D, d)
	}
}

// This test relies on deprecated functions in crypto/elliptic
// This is only included to make sure that sign/verify is compatible
// a known verified implementation.
func TestCompat(t *testing.T) {
	var cc = ec.NewCompatCurve(ec.DemoCurve25)
	var h = sha256.Sum256([]byte("you can't edit this"))
	var pk1, err = GenerateKey(ec.DemoCurve25, rand.Reader)
	assert.Nil(t, err)
	var pk2 = ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: cc,
			X:     big.NewInt(pk1.Pub.P.X),
			Y:     big.NewInt(pk1.Pub.P.Y),
		},
		D: big.NewInt(pk1.D),
	}

	// Sign with ecdsa
	r, s, err := ecdsa.Sign(rand.Reader, &pk2, h[:])

	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.NotNil(t, s)

	// Verify with Verify
	assert.True(t, Verify(pk1.Pub, r.Int64(), s.Int64(), h[:]))

	// Sign with Sign and verify with ecdsa
	h = sha256.Sum256([]byte("this is too immutable"))

	r1, s1, err := Sign(rand.Reader, pk1, h[:])

	assert.Nil(t, err)
	assert.True(t, ecdsa.Verify(&pk2.PublicKey,
		h[:],
		big.NewInt(r1),
		big.NewInt(s1)))
}
