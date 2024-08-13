package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

type AESConfig struct {
	Nonce []byte
	GCM   cipher.AEAD
}

func (c *AESConfig) Encrypt(plaintext []byte) string {
	ciphertext := c.GCM.Seal(nil, c.Nonce, plaintext, nil)
	return hex.EncodeToString(c.Nonce) + hex.EncodeToString(ciphertext)
}

func (c *AESConfig) Decrypt(ciphertextHex string) (string, error) {
	cipherData, _ := hex.DecodeString(ciphertextHex)
	nonceSize := 12
	nonce, ciphertext := cipherData[:nonceSize], cipherData[nonceSize:]
	plaintext, err := c.GCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func NewAESConfig() (*AESConfig, error) {
	block, err := aes.NewCipher([]byte("0E00894B1D18FFB84E1D2E0DBA5611BB"))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return &AESConfig{
		Nonce: nonce,
		GCM:   gcm,
	}, nil
}
