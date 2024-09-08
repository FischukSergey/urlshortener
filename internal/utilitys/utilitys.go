package utilitys

import (
	"math/rand"
	"strconv"
	"time"
)

// NewRandomString generates random string with given size.
func NewRandomString(size int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	b := make([]rune, size)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}
	return string(b)
}

// NewRandomID генератор случайного ID
func NewRandomID(size int) int {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []rune("0123456789")

	b := make([]rune, size)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}
	id, _ := strconv.Atoi(string(b))
	return id
}
