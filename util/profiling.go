package util

import (
	"crypto/md5"
	"crypto/rand"
)

func HeavyOperation() {
	for {
		token := make([]byte, 1024)
		rand.Read(token)
		md5.New().Write(token)
	}
}
