package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

// шифрование
func encryptData(plaintext []byte, key []byte) ([]byte, error) {
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

// дешифрование
func decryptData(encrypted []byte, key []byte) ([]byte, error) {
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

/*
// Клиент при добавлении данных:
key := deriveKey(masterPassword, salt)

// Подготовка данных
cardData := []byte(`{
    "number": "4111111111111111",
    "expiry": "12/25",
    "cvv": "123"
}`)

encrypted, err := encryptData(cardData, key)
if err != nil {
// Обработка ошибки
}

// Отправка на сервер
_, err = db.Exec(`
    INSERT INTO secrets (user_id, type, encrypted_data)
    VALUES ($1, $2, $3)`,
userID, "card", encrypted,
)


//Клиент при получении данных
var encrypted []byte
err := db.QueryRow(`
    SELECT encrypted_data
    FROM secrets
    WHERE id = $1 AND user_id = $2`,
secretID, userID,
).Scan(&encrypted)

decrypted, err := decryptData(encrypted, key)
if err != nil {
// Обработка ошибки
}

// Использование данных
var card struct {
	Number string `json:"number"`
	Expiry string `json:"expiry"`
	CVV    string `json:"cvv"`
}
json.Unmarshal(decrypted, &card)*/
