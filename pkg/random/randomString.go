package lib_random

import "crypto/rand"

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321_"

func RandomString(n int) string {

	// define 2 vars
	// the first is just a slice of rune that is n length
	// the second is a slice of rune seeded from random string source

	s, r := make([]rune, n), []rune(randomStringSource)

	// loop through the empty rune slice
	for i := range s {
		// generate a prime number based on the given bit length of r log2(n)+1
		// for example, 12 would return a bit-length of 4 so the prime number would be based on 4
		p, _ := rand.Prime(rand.Reader, len(r))

		// define 2 additional variables,
		// x is based on the Unit64 representation of p from above
		// y is based on the uint64 type case of the length of R (our []rune(randomString))
		x, y := p.Uint64(), uint64(len(r)) // note: uint64 here because we know it will not be negative

		// finally for the index of if in s which is just an empty slice of rune
		// choose a  rune from r where the index is the result of modulus operationx x%y

		s[i] = r[x%y]
	}

	// after we finish looping through the rune and assigning values to each index,
	// return the string
	return string(s)
}