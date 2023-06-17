package testutils

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	mathrand "math/rand"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func SeedRand() {
	var b [8]byte
	_, err := cryptorand.Read(b[:])
	if err != nil {
		panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	mathrand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + mathrand.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[mathrand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner generates a random owner name
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 100)
}

// RandomCurrency generates a random currency code
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)
	return currencies[mathrand.Intn(n)]
}
