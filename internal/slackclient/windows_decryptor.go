package slackclient

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type WindowsDecryptor struct{}

func (WindowsDecryptor) Decrypt(value, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create NewGCM: %w", err)
	}

	decryptedValue := make([]byte, 0)
	decryptedValue, err = gcm.Open(decryptedValue, value[:12], value[12:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	return decryptedValue, nil
}
