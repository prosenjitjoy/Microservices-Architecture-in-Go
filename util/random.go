package util

import (
	"math/rand"
	"strings"
	"time"
)

var random *rand.Rand

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[random.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + random.Int63n(max-min+1)
}
