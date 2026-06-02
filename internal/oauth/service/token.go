package service

import (
	"errors"
	"os"

	"github.com/fernet/fernet-go"
)

func cipher() (*fernet.Key, error) {
	return fernet.DecodeKey(os.Getenv("FERNET_KEY"))
}

func Encrypt(plaintext string) (string, error) {
	key, err := cipher()
	if err != nil {
		return "", err
	}
	token, err := fernet.EncryptAndSign([]byte(plaintext), key)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func Decrypt(ciphertext string) (string, error) {
	key, err := cipher()
	if err != nil {
		return "", err
	}
	plaintext := fernet.VerifyAndDecrypt([]byte(ciphertext), 0, []*fernet.Key{key})
	if plaintext == nil {
		return "", errors.New("invalid or expired token")
	}
	return string(plaintext), nil
}
