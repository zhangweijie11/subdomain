package util

import (
	"math/rand"
	"time"
)

func RandomStr(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rng.Intn(len(letterRunes))]
	}
	return string(b)
}
func RandInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Int63n(max-min) + min
}
