package helpers

import (
	"math/big"

	"go.dedis.ch/kyber/v3/util/random"
)

// RandString returns a random string of N chars
func RandString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bigN := big.NewInt(int64(len(letters)))
	b := make([]byte, n)
	r := random.New()
	for i := range b {
		x := int(random.Int(bigN, r).Int64())
		b[i] = letters[x]
	}
	return string(b)
}
