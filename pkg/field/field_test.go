package field

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	var f = NewFinite(97)
	var tests = []struct {
		i, j, r int64
	}{
		{1, 2, 3},
		{0, 0, 0},
		{0, 1, 1},
		{49, 51, 3},
		{96, 96, 95},
		{96, 1, 0},
		{96, 2, 1},
		{-40, 38, 95},
		{-40, -38, 19},
	}

	for _, tc := range tests {
		var r = f.Add(tc.i, tc.j)
		assert.Equal(t, tc.r, r, "Add failed")
	}
}

func TestMultiply(t *testing.T) {
	var f = NewFinite(89)
	var tests = []struct {
		i, j, r int64
	}{
		{1, 2, 2},
		{0, 0, 0},
		{49, 0, 0},
		{49, 51, 7},
		{86, 86, 9},
		{-1, 1, 88},
		{-1, -1, 1},
	}

	for _, tc := range tests {
		var r = f.Multiply(tc.i, tc.j)
		assert.Equal(t, tc.r, r, "Multiply failed")
	}

	f = NewFinite(996488947583)
	tests = []struct {
		i, j, r int64
	}{
		{-1, -1, 1},
		{996488947583, 31, 0},
		{996488947582, 99648897583, 896840050000},
	}

	for _, tc := range tests {
		var r = f.Multiply(tc.i, tc.j)
		assert.Equal(t, tc.r, r, "Multiply failed")
	}
}

func TestCanonicalize(t *testing.T) {
	var f = NewFinite(89)
	var tests = []struct {
		i, r int64
	}{
		{-1, 88},
		{157, 68},
		{0, 0},
		{89, 0},
	}

	for _, tc := range tests {
		var r = f.Canonicalize(tc.i)
		assert.Equal(t, tc.r, r, "Canonicalize failed")

		r = f.canonicalize(big.NewInt(tc.i))
		assert.Equal(t, tc.r, r, "canonicalize failed")
	}
}

func TestInverse(t *testing.T) {
	t.Run("inverse", func(t *testing.T) {
		var tests = []struct {
			i, j, f int64
		}{
			{13, 70, 101},
			{77, 21, 101},
			{1234, 112, 2003},
		}

		for _, tc := range tests {
			var f = NewFinite(tc.f)
			var r, _ = f.Inverse(tc.i)

			assert.Equal(t, tc.j, r, "Inverse failed")
		}
	})

	t.Run("non inversible", func(t *testing.T) {
		var tests = []struct {
			i, f int64
		}{
			{0, 101},
			{1410, 2773},
		}

		for _, tc := range tests {
			var f = NewFinite(tc.f)
			var r, err = f.Inverse(tc.i)

			assert.Equal(t, int64(0), r, "Inverse failed for %d", tc.f)
			assert.NotNil(t, err)
		}
	})

	t.Run("division", func(t *testing.T) {
		var f = NewFinite(101)
		var tests = []struct {
			i, j, r int64
		}{
			{77, 13, 37},
			{13, 77, 71},
			{50, 2, 25},
			{-1, 1, 100},
			{-1, -1, 100},
		}

		for _, tc := range tests {
			var r, _ = f.Inverse(tc.j)
			r = f.Multiply(tc.i, r)

			assert.Equal(t, tc.r, r, "Division failed")
		}
	})
}

func TestExponentiate(t *testing.T) {
	var f = NewFinite(179)
	var tests = []struct {
		i, j, r int64
	}{
		//		{1, 2, 1},
		//		{33, 0, 1},
		//		{33, 33, 160},
		{2, 4, 16},
		//		{2, 5, 32},
		//		{2, 8, 77},
		//		{10, 10, 141},
	}

	for _, tc := range tests {
		var r = f.Exponentiate(tc.i, tc.j)

		assert.Equal(t, tc.r, r, "Exponentiate failed")
	}
}

func TestSqrt(t *testing.T) {
	var f = NewFinite(11)
	var tests = []struct {
		i   int64
		r   []int64
		err bool
	}{
		{i: 1, r: []int64{1, 10}},
		{i: 2, err: true},
		{i: 3, r: []int64{5, 6}},
		{i: 4, r: []int64{2, 9}},
		{i: 5, r: []int64{4, 7}},
		{i: 6, err: true},
		{i: 7, err: true},
		{i: 8, err: true},
		{i: 9, r: []int64{3, 8}},
		{i: 10, err: true},
	}

	for _, tc := range tests {
		var found bool
		var s int64
		var err error

		s, err = f.Sqrt(tc.i)
		if tc.err == (err == nil) {
			t.Errorf("unexpected error for %d", tc.i)
		}

		if len(tc.r) == 0 {
			// tc.i is not a square
			continue
		}

		for _, ss := range tc.r {
			if s == ss {
				found = true
			}
		}

		if !found {
			t.Errorf("wrong root %d for %d", s, tc.i)
		}
	}
}
