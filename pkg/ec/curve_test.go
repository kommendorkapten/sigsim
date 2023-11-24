package ec

import (
	"testing"

	"github.com/kommendorkapten/sigsim/pkg/field"
	"github.com/stretchr/testify/assert"
)

func TestPoint(t *testing.T) {
	var tests = []struct {
		p1, p2 Point
		eq     bool
	}{
		{
			p1: Point{X: 123, Y: 345},
			p2: Point{X: 123, Y: 345},
			eq: true,
		},
		{
			p1: Point{X: 123, Y: 345},
			p2: Point{X: 123, Y: 3},
			eq: false,
		},
		{
			p1: Point{X: 1, Y: 345},
			p2: Point{X: 123, Y: 345},
			eq: false,
		},
		{
			p1: Point{X: 1, Y: 345},
			p2: Point{X: 123, Y: 3},
			eq: false,
		},
	}

	for _, tc := range tests {
		var eq = tc.p1.Equal(tc.p2)

		if eq != tc.eq {
			t.Errorf("equality test failed: %+v vs %+v",
				tc.p1, tc.p2)
		}
	}
}

func TestNewCurve(t *testing.T) {
	var tests = []struct {
		f  int64
		bs int
	}{
		{
			f:  5,
			bs: 3,
		},
		{
			f:  251,
			bs: 8,
		},
		{
			f:  25183,
			bs: 15,
		},
		{
			f:  57527,
			bs: 16,
		},
		{
			f:  3411011,
			bs: 22,
		},
		{
			f:  33480829,
			bs: 25,
		},
		{
			f:  33489583,
			bs: 25,
		},
		{
			f:  12864031,
			bs: 24,
		},
		{
			f:  114098711,
			bs: 27,
		},
		{
			f:  2095180127,
			bs: 31,
		},
	}

	for _, tc := range tests {
		c, err := NewCurve(field.NewFinite(tc.f), 2, 3)

		assert.Nil(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, tc.bs, c.BS, "wrong bitsize")
	}
}

func TestCurveAdd(t *testing.T) {
	t.Run("curve N 5 A 2 B 3", func(t *testing.T) {
		var c = Curve{
			F: field.NewFinite(5),
			A: 2,
			B: 3,
		}

		var tests = []struct {
			p, q, r Point
		}{
			{
				p: Point{X: 1, Y: 4},
				q: Point{X: 3, Y: 1},
				r: Point{X: 2, Y: 0},
			},
			{
				p: Point{X: 1, Y: 4},
				q: Point{X: 1, Y: 4},
				r: Point{X: 3, Y: 1},
			},
			{
				p: Point{X: 1, Y: 4},
				q: Point{X: 1, Y: 4},
				r: Point{X: 3, Y: 1},
			},
			{
				p: Point{X: 2, Y: 0},
				q: Point{X: 2, Y: 0},
				r: Point{Inf: true},
			},
		}

		for _, tc := range tests {
			var r = c.Add(tc.p, tc.q)

			assert.Equal(t, tc.r, r, "Add point failed")
		}
	})

	t.Run("curve N 2773 A 4 B 4", func(t *testing.T) {
		var c = Curve{
			F: field.NewFinite(2773),
			A: 4,
			B: 4,
		}

		var tests = []struct {
			p, q, r Point
		}{
			{
				p: Point{X: 1, Y: 3},
				q: Point{X: 1, Y: 3},
				r: Point{X: 1771, Y: 705},
			},
		}

		for _, tc := range tests {
			var r = c.Add(tc.p, tc.q)

			assert.Equal(t, tc.r, r, "Add point failed")
		}
	})

	t.Run("curve N 2773 A 0 B 73", func(t *testing.T) {
		var c = Curve{
			F: field.NewFinite(2773),
			A: 0,
			B: 73,
		}

		var tests = []struct {
			p, q, r Point
		}{
			{
				p: Point{X: 2, Y: 9},
				q: Point{X: 3, Y: 10},
				r: Point{X: 2769, Y: 2770},
			},
		}

		for _, tc := range tests {
			var r = c.Add(tc.p, tc.q)

			assert.Equal(t, tc.r, r, "Add point failed")
		}
	})

	t.Run("curve N 11 A 1 B 6", func(t *testing.T) {
		var c = Curve{
			F: field.NewFinite(11),
			A: 1,
			B: 6,
		}

		var tests = []struct {
			p, q, r Point
		}{
			{
				p: Point{X: 2, Y: 4},
				q: Point{X: 2, Y: 7},
				r: Point{Inf: true},
			},
		}

		for _, tc := range tests {
			var r = c.Add(tc.p, tc.q)

			assert.Equal(t, tc.r, r, "Add point failed")
		}
	})

	t.Run("curve N 25698457 A 2 B 3", func(t *testing.T) {
		var c = Curve{
			F: field.NewFinite(25698457),
			A: 2,
			B: 3,
		}

		var tests = []struct {
			p, q, r Point
		}{
			{
				p: Point{X: 11734117, Y: 14055369},
				q: Point{X: 11734117, Y: 14055369},
				r: Point{X: 6501286, Y: 10291724},
			},
		}

		for _, tc := range tests {
			var r = c.Add(tc.p, tc.q)

			assert.Equal(t, tc.r, r, "Add point failed")

			if !c.Valid(r) {
				panic(r)
			}
		}
	})
}

func TestScalarM(t *testing.T) {
	t.Run("curve N 5 A 2 B 3", func(t *testing.T) {
		var c = Curve{
			F: field.NewFinite(5),
			A: 2,
			B: 3,
		}

		var tests = []struct {
			p, r Point
			k    int64
		}{
			{
				p: Point{X: 1, Y: 4},
				r: Point{X: 1, Y: 4},
				k: 1,
			},
			{
				p: Point{X: 1, Y: 4},
				r: Point{X: 3, Y: 1},
				k: 2,
			},
			{
				p: Point{X: 1, Y: 4},
				r: Point{X: 2, Y: 0},
				k: 3,
			},
			{
				p: Point{X: 1, Y: 4},
				r: Point{X: 3, Y: 4},
				k: 4,
			},
		}

		for _, tc := range tests {
			var r = c.ScalarM(tc.k, tc.p)

			assert.Equal(t, tc.r, r, "Failed scalar")

			// Verify repeatable add gives the same answer
			r = Point{Inf: true}
			var ik int64
			for ik = 0; ik < tc.k; ik++ {
				r = c.Add(r, tc.p)
			}

			assert.Equal(t, tc.r, r, "Failed add")
		}
	})

	t.Run("curve N 2773 A 4 B 4", func(t *testing.T) {
		var c = Curve{
			F: field.NewFinite(2773),
			A: 4,
			B: 4,
		}

		var tests = []struct {
			p, r Point
			k    int64
		}{
			{
				p: Point{X: 1, Y: 3},
				r: Point{X: 1771, Y: 705},
				k: 2,
			},
			{
				p: Point{X: 1, Y: 3},
				r: Point{Inf: true},
				k: 0,
			},
		}

		for _, tc := range tests {
			var r = c.ScalarM(tc.k, tc.p)

			assert.Equal(t, tc.r, r, "Add point failed")
		}
	})

	t.Run("test commutative", func(t *testing.T) {
		var c *Curve
		var err error

		c, err = NewCurve(field.NewFinite(2767),
			-3 + 2767,
			4)
		assert.Nil(t, err)

		var g = c.RandomPoint()
		var d1, d2 int64

		// Generate two random d values (take the x coordinate from
		// two points
		d1 = c.RandomPoint().X
		for {
			d2 = c.RandomPoint().X

			if d1 != d2 {
				break
			}
		}
		var g1 = c.ScalarM(d1, c.ScalarM(d2, g))
		var g2 = c.ScalarM(d2, c.ScalarM(d1, g))

		assert.Equal(t, g1, g2)
	})
}

func TestPointsOnCurve(t *testing.T) {
	var c = Curve{
		F: field.NewFinite(5),
		A: 2,
		B: 3,
	}
	var exp = []Point{
		{Inf: true},
		{X: 1, Y: 1},
		{X: 1, Y: 4},
		{X: 2, Y: 0},
		{X: 3, Y: 1},
		{X: 3, Y: 4},
		{X: 4, Y: 0},
	}
	var got = c.Points()

	assert.ElementsMatch(t, exp, got, "wrong number of points")

	for _, e := range exp {
		var found = false

		for _, g := range got {
			if e == g {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("failed to find point %+v", e)
		}
	}

	c = Curve{
		F: field.NewFinite(101),
		A: 0,
		B: 73,
	}
	got = c.Points()
	assert.Equal(t, 102, len(got), "wrong number of points")

	c = Curve{
		F: field.NewFinite(71),
		A: -1,
		B: 0,
	}
	got = c.Points()
	assert.Equal(t, 72, len(got), "wrong number of points")
}

func TestOrder(t *testing.T) {
	t.Parallel()
	t.Run("curve_263_2_3", func(t *testing.T) {
		t.Parallel()
		var c = Curve{
			F: field.NewFinite(263),
			A: 2,
			B: 3,
		}
		// Test cardinality first
		var points = c.Points()
		assert.Equal(t, 270, len(points), "wrong number of points")

		// Pick a "good" point (i.e generator)
		var g = Point{X: 200, Y: 39}
		assert.True(t, c.Valid(g))
		var order = c.Order(g)
		var orderBG = c.OrderBG(g)
		assert.Equal(t, int64(270), order, "wrong order")
		assert.Equal(t, order, orderBG, "wrong order")

		// Now use a "bad" genertor (order 9)
		g = Point{X: 60, Y: 48}
		order = c.Order(g)
		orderBG = c.OrderBG(g)
		assert.True(t, c.Valid(g))
		assert.Equal(t, int64(9), order, "wrong order")
		assert.Equal(t, order, orderBG, "wrong order")

		// Enumerate all points and verify we get back to where
		// we start
		points = []Point{
			{X: 60, Y: 48},
			{X: 128, Y: 81},
			{X: 144, Y: 35},
			{X: 102, Y: 90},
			{X: 102, Y: 173},
			{X: 144, Y: 228},
			{X: 128, Y: 182},
			{X: 60, Y: 215},
			{Inf: true},
			{X: 60, Y: 48},
		}

		var acc = Point{Inf: true}
		for i, p := range points {
			var r = c.ScalarM(int64(i+1), g)
			acc = c.Add(acc, g)

			if r != acc {
				t.Errorf("%+v != %+v", r, acc)
			}

			if p != acc {
				t.Errorf("expected %+v, got %+v", p, acc)
			}
		}
	})

	t.Run("curve_15_0_7", func(t *testing.T) {
		t.Parallel()

		var c = Curve{
			F: field.NewFinite(17),
			A: 0,
			B: 7,
		}
		// Test cardinality first
		var points = c.Points()
		assert.Equal(t, 18, len(points), "wrong number of points")

		var tests = []struct{
			g Point
			r int64
		}{
			{
				g: Point{X: 15, Y: 13},
				r: 18,
			},
			{
				g: Point{X: 3, Y: 0},
				r: 2,
			},
			{
				g: Point{X: 5, Y: 9},
				r: 3,
			},
		}
		for _, tc := range tests {
			assert.True(t, c.Valid(tc.g))
			assert.Equal(t, tc.r, c.Order(tc.g))
			assert.Equal(t, tc.r, c.OrderBG(tc.g))
		}
	})

	// This test takes about 130s on a M1 CPU.
	// t.Run("curve_33489583_-3_3411011", func(t *testing.T) {
	// 	t.Parallel()
	// 	var c = Curve{
	// 		F: field.NewFinite(33489583),
	// 		A: -3,
	// 		B: 3411011,
	// 	}

	// 	// Pick a "good" point (i.e generator).
	// 	var g = Point{X: 12272011, Y: 8490180}
	// 	var order = c.Order(g)
	// 	var want int64 = 33480829
	// 	assert.Equal(t, want, order, "wrong order")
	// 	var orderBG = c.OrderBG(g)
	// 	assert.Equal(t, want, orderBG, "wrong order (bg)")
	// })
}

// This test takes aound 90s on a M1 CPU.
// func TestVerify(t *testing.T) {
// 	t.Parallel()
// 	assert.Nil(t, DemoCurve25.Verify())
// }
