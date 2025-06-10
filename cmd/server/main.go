package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"time"
)

var userKeys = make(map[string]string) // username -> public key

func main() {
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/protected", authMiddleware(protectedHandler))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Username  string `json:"username"`
		PublicKey string `json:"public_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Сохраняем публичный ключ
	userKeys[data.Username] = data.PublicKey
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			Username  string `json:"username"`
			Timestamp string `json:"timestamp"`
			Nonce     string `json:"nonce"`
			Signature string `json:"signature"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Проверяем временную метку
		t, err := time.Parse(time.RFC3339, data.Timestamp)
		if err != nil || time.Since(t).Abs() > 5*time.Minute {
			http.Error(w, "Invalid or expired timestamp", http.StatusUnauthorized)
			return
		}

		// Получаем публичный ключ
		pubKeyPEM, ok := userKeys[data.Username]
		if !ok {
			http.Error(w, "User not registered", http.StatusUnauthorized)
			return
		}

		// Парсим публичный ключ
		block, _ := pem.Decode([]byte(pubKeyPEM))
		if block == nil {
			http.Error(w, "Invalid public key format", http.StatusInternalServerError)
			return
		}

		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			http.Error(w, "Failed to parse public key", http.StatusInternalServerError)
			return
		}

		rsaPubKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			http.Error(w, "Not an RSA public key", http.StatusInternalServerError)
			return
		}

		// Проверяем подпись
		signature, err := base64.StdEncoding.DecodeString(data.Signature)
		if err != nil {
			http.Error(w, "Invalid signature encoding", http.StatusBadRequest)
			return
		}

		payload := []byte(fmt.Sprintf("%s|%s|%s", data.Username, data.Timestamp, data.Nonce))
		hashed := sha256.Sum256(payload)

		if err := rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashed[:], signature); err != nil {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		// Аутентификация успешна
		next.ServeHTTP(w, r)
	}
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Access granted to protected resource!"))
}
