package main

import (
	"strings"
	"testing"

	"extremeWorkload.com/daytrader/lib/security"
)

func doTest(t *testing.B, message string) {
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		encrypted, _ := security.Encrypt(message)
		_, _ = security.Decrypt(encrypted)
	}
}
func BenchmarkEncryptionSmall(t *testing.B) {
	security.InitCryptoKey()
	message := strings.Repeat("1234567890", 10)
	doTest(t, message)
}

func BenchmarkEncryptionMed(t *testing.B) {
	security.InitCryptoKey()
	message := strings.Repeat("1234567890", 100)
	doTest(t, message)
}

func BenchmarkEncryptionLarge(t *testing.B) {
	security.InitCryptoKey()
	message := strings.Repeat("1234567890", 5000)
	doTest(t, message)
}
