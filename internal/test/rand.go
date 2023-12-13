package test

import (
	"math/rand"
	"time"
)

func RandomName() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}
