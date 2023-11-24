package math

// PrimeFactors returns an ordered list of the prime factors using
// a very naive algorithm.
func PrimeFactors(n int64) []int64 {
	var pfs []int64

	// The easy one first
	for n % 2 == 0 {
		pfs = append(pfs, 2)
		n /= 2
	}

	// n is odd, so skip even numbers
	for i := int64(3); i*i <= n; i = i + 2 {
		// prime factors can be repeated
		for n%i == 0 {
			pfs = append(pfs, i)
			n /= i
		}
	}

	// if n is prime, return it
	if n > 2 {
		pfs = append(pfs, n)
	}

	return pfs
}
