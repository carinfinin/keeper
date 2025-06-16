package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

type Config struct {
	Addr           string `json:"address"`
	DBPath         string `json:"database_dsn"`
	IsTLS          bool   `json:"enable_https"`
	LogLevel       string `json:"log_level"`
	PrivateKeyPath string `json:"path_private_key"`
	PublicKeyPath  string `json:"path_public_key"`
	PrivateKey     *rsa.PrivateKey
	PublicKey      *rsa.PublicKey
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`

	JWTKeyID             string        `json:"jwt_key_id"`
	JWTAudience          string        `json:"app_name"`
	AccessTokenDuration  time.Duration `json:"access_token_duration"`
	RefreshTokenDuration time.Duration `json:"refresh_token_duration"`
}

func New() *Config {
	var configPath string
	flag.StringVar(&configPath, "c", "config.json", "config file path")
	flag.Parse()

	if configPath == "" {
		log.Fatal("not found config path")
	}
	c, err := configRead(configPath)
	if err != nil {
		log.Fatal("error read file config : ", err)
	}
	return c
}

func configRead(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err = json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	//privatePEM, err := os.ReadFile(cfg.PrivateKeyPath)
	//if err != nil {
	//	log.Fatal("Failed to read private key:", err)
	//}

	//publicPEM, err := os.ReadFile(cfg.PublicKeyPath)
	//if err != nil {
	//	log.Fatal("Failed to read public key:", err)
	//}

	//cfg.PrivateKey, cfg.PublicKey, err = loadKeys(string(privatePEM), string(publicPEM))
	//if err != nil {
	//	log.Fatal(err)
	//}

	return &cfg, nil
}

func loadKeys(private, public string) (*rsa.PrivateKey, *rsa.PublicKey, error) {

	// PublicKey
	pubBlock, _ := pem.Decode([]byte(public))
	if pubBlock == nil {
		return nil, nil, errors.New("failed to parse PEM public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("not an RSA public key")
	}

	// PrivateKey
	privBlock, _ := pem.Decode([]byte(private))
	if privBlock == nil {
		return nil, nil, errors.New("failed to parse PEM private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("not an RSA public key")
	}

	return rsaPrivateKey, rsaPubKey, nil
}
