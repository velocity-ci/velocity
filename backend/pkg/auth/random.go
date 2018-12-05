package auth

// From https://gist.github.com/shahaya/635a644089868a51eccd6ae22b2eb800

import "crypto/rand"

// GenerateRandomBytes returns securely generated random bytes.
func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

// RandomString returns a securely generated random string.
func RandomString(n int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes := generateRandomBytes(n)
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes)
}
