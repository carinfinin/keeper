package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const keyBits = 4096

func main() {
	register("http://localhost:8080/register", "alex")
}

func register(apiURL, username string) {
	// Генерация ключей
	privKeyPath, pubKeyPath := getKeyPaths(username)
	if err := generateKeyPair(privKeyPath, pubKeyPath); err != nil {
		log.Fatalf("Key generation failed: %v", err)
	}

	// Чтение публичного ключа
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatalf("Failed to read public key: %v", err)
	}

	payload := map[string]string{
		"username":   username,
		"public_key": string(pubKey),
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Registration failed: %s", body)
	}

	fmt.Println("Registration successful")
}

func makeRequest(apiURL, username string) {
	// Загружаем приватный ключ
	privKeyPath, _ := getKeyPaths(username)
	privateKey, err := loadPrivateKey(privKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем подписываемые данные
	timestamp := time.Now().UTC().Format(time.RFC3339)
	nonce := generateNonce(16)
	payload := []byte(fmt.Sprintf("%s|%s|%s", username, timestamp, nonce))

	// Подписываем данные
	signature, err := signData(privateKey, payload)
	if err != nil {
		log.Fatal(err)
	}

	// Формируем запрос
	reqBody := map[string]string{
		"username":  username,
		"timestamp": timestamp,
		"nonce":     nonce,
		"signature": base64.StdEncoding.EncodeToString(signature),
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\nBody: %s\n", resp.StatusCode, body)
}

func signData(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func generateNonce(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// Функции generateKeyPair и getKeyPaths аналогичны предыдущим примерам
// Пути к ключам
func getKeyPaths(username string) (string, string) {
	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh", "myapp")
	os.MkdirAll(sshDir, 0700)

	privKey := filepath.Join(sshDir, fmt.Sprintf("id_%s", username))
	pubKey := privKey + ".pub"

	return privKey, pubKey
}

// Генерация ключей
func generateKeyPair(privPath, pubPath string) error {
	// Проверка существования ключей
	if _, err := os.Stat(privPath); err == nil {
		return fmt.Errorf("private key already exists: %s", privPath)
	}

	// Генерация ключа
	privateKey, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return err
	}

	// Сохранение приватного ключа
	privFile, err := os.OpenFile(privPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer privFile.Close()

	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(privFile, privBlock); err != nil {
		return err
	}

	// Генерация публичного ключа
	pubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	// Сохранение публичного ключа
	pubKeyBytes := ssh.MarshalAuthorizedKey(pubKey)
	return os.WriteFile(pubPath, pubKeyBytes, 0644)
}
