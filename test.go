package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func main() {
	sum := sha256.Sum256([]byte("2325d724e353e5a1df4554005a5fa0a196b1be9650795e211805fa59d90c0d7b"))

	fmt.Println(hex.EncodeToString(sum[:]))
}
