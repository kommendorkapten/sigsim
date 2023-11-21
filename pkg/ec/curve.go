// Package ec implements arithmetic on an elliptic curve over a finite field.
package ec

import (
	"fmt"
	"math/big"
	"time"

	"github.com/kommendorkapten/sigsim/pkg/field"
)

// Point represents a point on a curve.
// If the point is the identity element, Inf is set to true.
type Point struct {
	X   int64
	Y   int64
	Inf bool
}

// Equal test if two points are qual.
func (p Point) Equal(cmp Point) bool {
	if p.X != cmp.X {
		return false
	}

	if p.Y != cmp.Y {
		return false
	}

	return p.Inf == cmp.Inf
}

// Curve represents an elliptic curve over a finite field satisfying
// the following equation Y^2 = X^3 + ax + b
type Curve struct {
	F  *field.Finite
	A  int64 // A parameter
	B  int64 // B parameter
	G  Point // Generator point
	N  int64 // Order of the generator point
	BS int   // Bitsize of the underlying field
}

// DemoCurve25 is a simple curve over a field of bitlength 25.
var DemoCurve25 = &Curve{
	F: field.NewFinite(33489583),
	A: -3 + 33489583,
	B: 3411011,
	G: Point{
		X: 12272011,
		Y: 8490180,
	},
	N:  33480829,
	BS: 25,
}

// DemoCurve63 is a simple curve over a field of bitlength 63.
var DemoCurve63 = &Curve{
	F: field.NewFinite(8527849010035426663),
}

// NewCurve returns a Weierstrass curve over the finite field, using the
// provided a and b parameters. A _should_ be -3. If the curve's
// discriminant is zero, an error is returned.
func NewCurve(f *field.Finite, a, b int64) (*Curve, error) {
	if !f.Element(a) {
		return nil, fmt.Errorf(
			"a: %d is not an element of provided field",
			a,
		)
	}

	if !f.Element(b) {
		return nil, fmt.Errorf(
			"b: %d is not an element of provided field",
			b,
		)
	}

	var bs int

	for i := field.BitLength - 1; i > 0; i-- {
		if (f.P() & (1 << i)) != 0 {
			bs = i + 1

			break
		}
	}

	var c = &Curve{
		F:  f,
		A:  a,
		B:  b,
		BS: bs,
	}
	// The discriminant must be non-zero
	var discriminant = c.Discriminant()
	if discriminant.BitLen() == 0 {
		return nil, fmt.Errorf(
			"provided parameters are not valid, a:%d b:%d",
			a, b,
		)
	}

	return c, nil
}

// Discriminant computes the discriminant of the curve.
// nolint: lll
// nolint: revive
// Discriminant is 4a^3 + 27b^2
// See https://math.stackexchange.com/questions/1653368/discriminant-of-elliptic-curves
// for more details.
func (c *Curve) Discriminant() *big.Int {
	var discriminant big.Int
	var p = big.NewInt(c.A)
	p.Exp(p, big.NewInt(3), nil)
	discriminant.Mul(big.NewInt(4), p)
	p.SetInt64(c.B)
	p.Exp(p, big.NewInt(2), nil)
	p.Mul(big.NewInt(27), p)
	discriminant.Add(&discriminant, p)

	return &discriminant
}

// Verify verifies all the parameters of the curve.
func (c *Curve) Verify() error {
	// Is the underlying field a prime
	var p = big.NewInt(c.F.P())
	if !p.ProbablyPrime(256) {
		return fmt.Errorf("field order %d is not prime", c.F.P())
	}

	if c.Discriminant().BitLen() == 0 {
		return fmt.Errorf("discriminant is non-zero")
	}

	// Is generator point on curve
	if !c.Valid(c.G) {
		return fmt.Errorf("generator point is not on curve")
	}

	// Verify order of the point
	var order = c.Order(c.G)
	if order != c.N {
		return fmt.Errorf("invalid order for generator point %d, expected %d",
			order, c.N)
	}

	p = big.NewInt(c.N)
	if !p.ProbablyPrime(256) {
		return fmt.Errorf("curve order %d is not prime", c.N)
	}

	// Verify the bitsize of the order
	if p.BitLen() != c.BS {
		return fmt.Errorf("bitlength %d for curver order does not match %d",
			p.BitLen(), c.BS)
	}

	return nil
}

// Points calculates and returns all points on the curve.
// This function can take long time to finish, use with caution.
func (c *Curve) Points() []Point {
	var points = map[Point]struct{}{}
	var canSquareRoot = (c.F.P() % 4) == 3

	// Add the infinity point
	points[Point{Inf: true}] = struct{}{}

	var x, y int64
	for x = 0; x < c.F.P(); x++ {
		if canSquareRoot {
			var err error
			if y, err = c.Y(x); err == nil {
				var p = Point{X: x, Y: y}
				var valid = c.Valid(p)


				if valid {
					// -y is also a valid point, but as
					// -y is equal to (-y + n) mod n
					// we don't need to add it as it will
					// be added during the exhaustive
					// search
					points[p] = struct{}{}
				}
			}

			continue
		}

		for y = 0; y < c.F.P(); y++ {
			var p = Point{X: x, Y: y}
			var valid = c.Valid(p)
			if valid {
				points[p] = struct{}{}
			}
		}
	}

	var res = make([]Point, 0, len(points))
	for p := range points {
		res = append(res, p)
	}

	return res
}

// Y returns the (positive) y coordinate on the curve for a given x
// coordinate. If x is not on the curve, an error is returned.
func (c *Curve) Y(x int64) (int64, error) {
	// Y^2 = X^3 + ax + b
	var y int64
	var rhs = c.F.Exponentiate(x, 3)
	var err error

	rhs = c.F.Add(rhs, c.F.Multiply(c.A, x))
	rhs = c.F.Add(rhs, c.B)

	y, err = c.F.Sqrt(rhs)
	if err != nil {
		var msg = "x: %d does not appear to be on the curve %w"

		return 0, fmt.Errorf(msg, x, err)
	}

	return y, nil
}

// Add two points together and returns the resulting point.
// If p and q are the some point, p is doubled.
func (c *Curve) Add(p, q Point) Point {
	var r Point
	var m int64

	if p.Inf {
		return q
	}

	if q.Inf {
		return p
	}

	if p.Equal(q) {
		var inv int64
		var err error

		m = c.F.Multiply(3, c.F.Multiply(p.X, p.X))
		m = c.F.Add(m, c.A)
		inv = c.F.Multiply(2, p.Y)

		inv, err = c.F.Inverse(inv)
		if err != nil {
			// infinite slope
			return Point{Inf: true}
		}

		m = c.F.Multiply(m, inv)
	} else {
		var inv int64
		var err error

		m = c.F.Add(q.Y, -p.Y)
		inv = c.F.Add(q.X, -p.X)
		inv, err = c.F.Inverse(inv)
		if err != nil {
			// infinite slope
			return Point{Inf: true}
		}

		m = c.F.Multiply(m, inv)
	}

	r.X = c.F.Multiply(m, m)
	r.X = c.F.Add(r.X, -p.X)
	r.X = c.F.Add(r.X, -q.X)

	r.Y = c.F.Add(p.X, -r.X)
	r.Y = c.F.Multiply(r.Y, m)
	r.Y = c.F.Add(r.Y, -p.Y)

	return r
}

// ScalarM calculates the scalar multiplication of a point.
// This function is the meat of ec cryptography. Solving the inverse:
// xP = P' for a known P and P' is the discrete logarithm problem for
// elliptic curves.
func (c *Curve) ScalarM(k int64, p Point) Point {
	var r = Point{Inf: true}

	for b := 0; b < field.BitLength; b++ {
		if (k & int64(1<<b)) != 0 {
			r = c.Add(r, p)
		}

		p = c.Add(p, p)
	}

	return r
}

// Valid returns true if the provided point is a valid curve point.
func (c *Curve) Valid(p Point) bool {
	var lhs = c.F.Exponentiate(p.Y, 2)
	var rhs = c.F.Exponentiate(p.X, 3)
	rhs = c.F.Add(rhs, c.F.Multiply(c.A, p.X))
	rhs = c.F.Add(rhs, c.B)
	var valid = lhs == rhs

	return valid
}

// Order calculates the order for the provided point.
// The order is the cardinality of the set of points that which can be
// reached by multiplying p with a scalar.
func (c *Curve) Order(p Point) int64 {
	var order int64
	var cp Point
	var pp = p

	start := time.Now()
	for {
		order++

		cp = c.Add(pp, p)
		if !c.Valid(cp) {
			panic(cp)
		}
		if cp == p {
			// We reached our starting point
			return order
		}

		pp = cp

		// 2*P may overflow if P is large.
		var z = big.NewInt(c.F.P())
		z.Lsh(z, 2)
		var o = big.NewInt(order)
		if o.Cmp(z) > 0 {
			fmt.Println(o.Cmp(z))
			fmt.Println(c.F.P())
			fmt.Println(order)
			panic(order)
		}

		if (order % 1000000) == 0 {
			dur := time.Since(start)
			searched := float64(order) / float64(c.F.P())
			fmt.Printf("Finished %f in %s\n", searched, dur)
			start = time.Now()
		}
	}
}

// Count the number of points
// is all points on a curve equal? i.e do they have the same order?
func (c *Curve) ScoofsOrder() int64 {
	return 0
}
