// Package security https://golang.org/pkg/crypto/cipher/#example_NewGCM_decrypt
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var key []byte

const defaultKeyFilePath = "dev_key.key"

// InitCryptoKey sets the key for security library using AES_256_KEY env var
func InitCryptoKey() error {
	keyFilePath := os.Getenv("AES_256_KEY_FILE")
	if keyFilePath == "" {
		keyFilePath = defaultKeyFilePath
	}
	var err error
	key, err = ioutil.ReadFile(keyFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	if len(key) != 32 {
		panic("AES Key is the wrong size")
	}
	return nil

}

// Encrypt performs GCM encryption
func Encrypt(text string) ([]byte, error) {
	if key == nil {
		return nil, errors.New("AES key was not initialized")
	}
	plaintext := []byte(text)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt performs GCM decryption
func Decrypt(ciphertext []byte) (string, error) {
	if key == nil {
		return "", errors.New("AES key was not initialized")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()

	if len(ciphertext) < nonceSize {
		return "", errors.New("Malformed ciphertext")
	}

	plaintext, err := aesgcm.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
