package ec

import (
	"crypto/elliptic"
	"math/big"
)

// CompatCurve implements the Curve interface in crypto/elliptic
// The usage of this is deprecated but, this is only to test compatibility.
type CompatCurve struct {
	cp *elliptic.CurveParams
}

var _ elliptic.Curve = &CompatCurve{}

// NewCompatCurve wraps curve as elliptic.Curve.
func NewCompatCurve(c *Curve) *CompatCurve {
	var cc = CompatCurve{
		cp: &elliptic.CurveParams{
			P:       big.NewInt(c.F.P()),
			N:       big.NewInt(c.N),
			B:       big.NewInt(c.B),
			Gx:      big.NewInt(c.G.X),
			Gy:      big.NewInt(c.G.Y),
			BitSize: c.BS,
			Name:    "CompatCurve",
		},
	}

	return &cc
}

// Params returns the curve parameters.
func (c *CompatCurve) Params() *elliptic.CurveParams {
	return c.cp
}

// IsOnCurve returns true if the point is on the curve.
func (c *CompatCurve) IsOnCurve(x, y *big.Int) bool {
	return c.cp.IsOnCurve(x, y)
}

// Add two points on the curve.
func (c *CompatCurve) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	return c.cp.Add(x1, y1, x2, y2)
}

// Double a single point on the curve.
func (c *CompatCurve) Double(x, y *big.Int) (*big.Int, *big.Int) {
	return c.cp.Double(x, y)
}

// ScalarMult multiplies the point by a scalar.
func (c *CompatCurve) ScalarMult(x, y *big.Int, k []byte) (*big.Int, *big.Int) {
	return c.cp.ScalarMult(x, y, k)
}

// ScalarBaseMult multiplies the generator point by the scalar.
func (c *CompatCurve) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	return c.cp.ScalarBaseMult(k)
}
