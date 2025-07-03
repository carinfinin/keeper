// keystore хранит ключи шифрует и дешифрует данные.
package keystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/carinfinin/keeper/internal/crypted"
	"io"
	"log"
	"os"
	"path/filepath"
)

// KeyStorage.
type KeyStorage struct {
	EncryptedKey string `json:"encrypted_key"`
	KeyHash      string `json:"key_hash"`
}

// getStoragePath.
func getStoragePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(configDir, "keystore.bin")
}

// SaveDerivedKey зосдаёт ключ шифрования по паролю и соли с сервера.
func SaveDerivedKey(password, serverSalt string) error {
	cryptoKey := crypted.DeriveKey(password, serverSalt)

	// Создаем хеш ключа для проверки
	keyHash := sha256.Sum256(cryptoKey)

	storage := KeyStorage{
		EncryptedKey: base64.StdEncoding.EncodeToString(cryptoKey),
		KeyHash:      base64.StdEncoding.EncodeToString(keyHash[:]),
	}

	return saveEncrypted(storage)
}

func saveEncrypted(ks KeyStorage) error {
	path := getStoragePath()

	data, err := json.Marshal(ks)
	if err != nil {
		return fmt.Errorf("failed to marshal key storage: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create keystore file: %w", err)
	}
	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("failed to write keystore data: %w", err)
	}

	if err = f.Sync(); err != nil {
		return fmt.Errorf("failed to sync keystore file: %w", err)
	}

	return nil
}

func encryptWithPassword(data []byte, password string) ([]byte, error) {
	k := sha256.Sum256([]byte(password))
	block, err := aes.NewCipher(k[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

func decryptWithPassword(ciphertext []byte, password string) ([]byte, error) {
	// 1. Получаем ключ из пароля (аналогично шифрованию)
	k := sha256.Sum256([]byte(password))

	// 2. Создаем AES-блок
	block, err := aes.NewCipher(k[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	// 3. Инициализируем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 4. Проверяем минимальную длину ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// 5. Разделяем nonce и зашифрованные данные
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 6. Дешифруем данные
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// GetDerivedKey возвращает ключ шифрования.
func GetDerivedKey() ([]byte, error) {
	path := getStoragePath()
	// Читаем файл с зашифрованным ключом
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Декодируем JSON структуру
	var storage KeyStorage
	if err := json.Unmarshal(data, &storage); err != nil {
		return nil, fmt.Errorf("failed to parse key file: %w", err)
	}

	// Декодируем base64 ключ
	cryptoKey, err := base64.StdEncoding.DecodeString(storage.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	// Дополнительная проверка целостности
	storedHash, err := base64.StdEncoding.DecodeString(storage.KeyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid key hash format: %w", err)
	}

	computedHash := sha256.Sum256(cryptoKey)
	if !bytes.Equal(computedHash[:], storedHash) {
		return nil, errors.New("key data corrupted - hash mismatch")
	}

	return cryptoKey, nil
}
