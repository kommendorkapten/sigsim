// Package ec implements arithmetic on an elliptic curve over a finite field.
package ec

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/kommendorkapten/sigsim/pkg/field"
	smath "github.com/kommendorkapten/sigsim/pkg/math"
)

var Parallel int64 = 2

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

func (c *Curve) String() string {
	return fmt.Sprintf("%d %d %d %+v %d %d",
		c.F.P(),
		c.A,
		c.B,
		c.G,
		c.N,
		c.BS,
	)
}

// DemoCurve25 is a simple curve over a field of bitlength 25.
// Count points takes 43 seconds
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
					// both (x, y) and (x, -y) are valid
					points[p] = struct{}{}
					if p.Y > 0 {
						p.Y = -p.Y + c.F.P()
						points[p] = struct{}{}
					}
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

// CountPoints using baby-step giant-step
// https://en.wikipedia.org/wiki/Counting_points_on_elliptic_curves
func (c *Curve) CountPoints() int64 {
	var f = math.Sqrt(float64(c.F.P()))
	var sq = int64(math.Ceil(f))
	var nMin = c.F.P() + 1 - 2 * sq
	var nMax = c.F.P() + 1 + 2 * sq

	var cnt = 0
	fmt.Printf("(%d, %d)\n", nMin, nMax)
	for {
		var p = c.RandomPoint()
		var mp = c.OrderBG(p)
		var n []int64

		cnt++
		fmt.Printf("Testing point %d\n", cnt)
		for c := nMin; c <= nMax; c++ {
			if (c % mp) == 0 {
				n = append(n, c)
			}
		}

		if len(n) == 1 {
			return n[0]
		}
	}
}

// RandomPoint returns a random point on the curve.
func (c *Curve) RandomPoint() Point {
	var x *big.Int
	var p Point
	var err error

	if (c.F.P() % 4) != 3 {
		// Can not calculate square root on this prime field.
		panic(c.F.P())
	}

	for {
		x, err = rand.Int(rand.Reader, big.NewInt(c.F.P()))
		if err != nil {
			panic(err)
		}
		p.X = x.Int64()
		p.Y, err = c.Y(p.X)
		if err != nil {
			continue
		}

		if c.Valid(p) {
			return p
		}
	}

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
	if p.Inf {
		return true
	}

	var lhs = c.F.Exponentiate(p.Y, 2)
	var rhs = c.F.Exponentiate(p.X, 3)
	rhs = c.F.Add(rhs, c.F.Multiply(c.A, p.X))
	rhs = c.F.Add(rhs, c.B)
	var valid = lhs == rhs

	return valid
}

func (c *Curve) Valid2(p Point) bool {
	if p.Inf {
		return true
	}

	//var lhs = c.F.Exponentiate(p.Y, 2)
	//var rhs = c.F.Exponentiate(p.X, 3)
	var lhs int64
	var rhs int64
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
	var cnt int64
	var cp Point
	var pp = p

	start := time.Now()
	for {
		order++

		cp = c.Add(pp, p)
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

	fmt.Println(cnt)
	return 0
}

type job struct{
	start, stop int64
	ctx context.Context
	c chan jobRes
	pj []Point
	q Point
	p Point
}

type jobRes struct {
	mp, j int64
}

// OrderBG computes the order of a point on the curve using
// Baby-step giant-step
func (c *Curve) OrderBG(p Point) int64 {
	// p^-4
	var f = math.Sqrt(math.Sqrt(float64(c.F.P())))
	var m = int64(math.Ceil(f + 0.5)) + 1
	var pj = make([]Point, m)
	var q = c.ScalarM(c.F.P() + 1, p)
	var mp int64 = -1

	// Precompute pj = j * p for j in [0, m]
	for j := int64(0); j < m; j++ {
		pj[j] = c.ScalarM(j, p)
	}

	// Found a point pj that satisfies
	// (q + k*2*m*p).X = pj.X
	var chunk = c.F.P() / Parallel
	var res = make(chan jobRes, 1)

	ctx, stop := context.WithCancel(context.Background())

	for i := int64(0); i < Parallel; i++ {
		var job = job{
			start: i * chunk,
			stop: (i + 1) * chunk,
			ctx: ctx,
			c: res,
			pj: pj,
			q: q,
			p: p,
		}

		// Let the last one do a bit of extra work
		if i == (Parallel - 1) {
			job.stop = c.F.P()
		}
		fmt.Println("Dispatching job")
		go c.approxOrder(&job)
	}

	var jr = <- res
	// Got a value, stop all goroutines
	stop()

	// (mp +/- j)*p is now the identity element
	var id Point
	id = c.ScalarM(jr.mp + jr.j, p)
	if id.Inf {
		mp = jr.mp + jr.j
	} else {
		mp = jr.mp - jr.j
	}

	// mp may be a composite, find the minimal value satisfying
	// mp * p = identity element
	var pfs = smath.PrimeFactors(mp)
	for i := 0; i < len(pfs); {
		var s = mp / pfs[i]

		// Residual is smaller than current prime factor
		if s == 0 {
			break
		}

		id = c.ScalarM(s, p)
		if id.Inf {
			mp = s
		} else {
			i++
		}
	}
	// mp is now the order of point p
	return mp
}

func (c *Curve) approxOrder(job *job) {
	var cnt = 0
	var m = int64(len(job.pj))

	var start = time.Now()
	fmt.Printf("start job %d -> %d\n", job.start, job.stop)
	for k := job.start; k < job.stop; k++ {
		var cand Point

		cand = c.Add(job.q, c.ScalarM(k * 2 * m, job.p))

		for j := int64(0); j < m; j++ {
			if cand.X == job.pj[j].X {
				var mp = c.F.P() + 1 + (k * 2 * m)

				fmt.Printf("done(%d): found %d %d\n", k, mp, j)
				// We found a valid point
				job.c <- jobRes{
					mp: mp,
					j: j,
				}
				return
			}
		}

		cnt++
		if (cnt % 10000) == 0 {
			if job.ctx.Err() != nil {
				// Context is cancelled
				return
			}

			fmt.Printf("Tried %d points in %s\n",
				cnt,
				time.Since(start),
			)
			start = time.Now()
		}
	}
	fmt.Printf("WTF, finished job\n")
}
