package crypted

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
)

// EncryptData шифрование
func EncryptData(plaintext []byte, key []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ciphertext...), nil
}

// DecryptData дешифрование
func DecryptData(encrypted []byte, key []byte) ([]byte, error) {
	if len(encrypted) < 28 { // 12 nonce + 16 tag + min 1 byte data
		return nil, fmt.Errorf("invalid ciphertext length")
	}

	nonce := encrypted[:12]
	ciphertext := encrypted[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonce, ciphertext, nil)
}

// GenerateSalt генерирует соль.
func GenerateSalt() (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// DeriveKey генерирует 32-байтный ключ из пароля и соли
func DeriveKey(password, salt string) []byte {
	return pbkdf2.Key(
		[]byte(password),
		[]byte(salt),
		100_000,
		32,
		sha256.New,
	)
}
