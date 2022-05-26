package slackclient

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"

	"golang.org/x/crypto/pbkdf2"
)

type UnixCookieDecryptor struct {
	rounds int
}

func (d UnixCookieDecryptor) Decrypt(value, key []byte) ([]byte, error) {
	dk := pbkdf2.Key(key, []byte("saltysalt"), d.rounds, 16, sha1.New)

	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, 16)
	for i := range iv {
		iv[i] = ' '
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(value, value)

	bytesToStrip := int(value[len(value)-1])

	return value[:len(value)-bytesToStrip], nil
}
