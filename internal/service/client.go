package service

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/store/models"
	"golang.org/x/term"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/carinfinin/keeper/internal/crypted"
	"github.com/carinfinin/keeper/internal/keystore"
	"github.com/carinfinin/keeper/internal/store/storesqlite"
)

type KeeperService struct {
	db  *sql.DB
	cfg *clientcfg.Config
}

func NewClientService(cfg *clientcfg.Config) (*KeeperService, error) {
	db, err := storesqlite.InitDB(cfg.DBPAth)
	if err != nil {
		return nil, err
	}
	return &KeeperService{db: db, cfg: cfg}, nil
}

func (s *KeeperService) Close() error {
	return s.db.Close()
}

func (s *KeeperService) GetDecryptedItem(ctx context.Context, uid string) (*models.Item, error) {
	key, err := keystore.GetDerivedKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	item, err := storesqlite.GetItem(ctx, s.db, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	decrypted, err := crypted.DecryptData(item.Data, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	item.Data = decrypted

	return item, nil
}

func (s *KeeperService) GetItem(ctx context.Context, uid string) (*models.Item, error) {
	return storesqlite.GetItem(ctx, s.db, uid)
}

func (s *KeeperService) RefreshTokens(ctx context.Context, refresh string) (*models.AuthResponse, error) {
	var rt struct {
		RefreshToken string `json:"token"`
	}
	rt.RefreshToken = refresh
	b, err := json.Marshal(rt)
	req, err := http.NewRequest(http.MethodPost, s.cfg.BaseURL+"/api/refresh", bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("Keeper/%s (*%s)", s.cfg.Version, s.cfg.DocsURL))
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer response.Body.Close()
	rb, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var t models.AuthResponse
	err = json.Unmarshal(rb, &t)
	if err != nil {
		return nil, err
	}

	err = storesqlite.SaveTokens(ctx, s.db, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *KeeperService) SaveFile(ctx context.Context, outputDir string, item *models.Item) (string, error) {
	key, err := keystore.GetDerivedKey()
	if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}

	decrypted, err := crypted.DecryptData(item.Data, key)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	var fileData models.File
	if err := json.Unmarshal(decrypted, &fileData); err != nil {
		return "", fmt.Errorf("Ошибка разбора файла: %v", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("Ошибка создания директории: %v", err)
	}

	outputPath := filepath.Join(outputDir, fileData.Name)

	if err := os.WriteFile(outputPath, fileData.Content, 0644); err != nil {
		return "", fmt.Errorf("Ошибка сохранения файла: %v", err)
	}
	return outputPath, nil
}

func (s *KeeperService) AddDecryptedItem(ctx context.Context, item *models.Item, data []byte) error {
	key, err := keystore.GetDerivedKey()
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}

	encrypted, err := crypted.EncryptData(data, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	item.Data = encrypted

	err = storesqlite.SaveItem(ctx, s.db, item)
	if err != nil {
		return fmt.Errorf("failed to encrypt write bd: %w", err)
	}
	return nil
}

func (s *KeeperService) GetDecryptedItems(ctx context.Context) ([]*models.Item, error) {
	key, err := keystore.GetDerivedKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	items, err := storesqlite.GetItems(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	for _, item := range items {
		decrypted, err := crypted.DecryptData(item.Data, key)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt: %w", err)
		}
		item.Data = decrypted
	}

	return items, nil
}

func (s *KeeperService) DeleteItem(ctx context.Context, uid string) error {

	return storesqlite.DeleteItem(ctx, s.db, uid)
}

func (s *KeeperService) GetLastChanges(ctx context.Context, lastSync time.Time) ([]*models.Item, error) {
	return storesqlite.GetLastItems(ctx, s.db, lastSync)
}

func (s *KeeperService) MergeLastChanges(ctx context.Context, items []*models.Item) error {
	return nil
}

func (s *KeeperService) UpdateItem(ctx context.Context, item *models.Item, data []byte) error {

	key, err := keystore.GetDerivedKey()
	if err != nil {
		return fmt.Errorf("Ошибка полученииия ключа шифрованиия: %v\n", err)
	}

	encrypted, err := crypted.EncryptData(data, key)
	if err != nil {
		return fmt.Errorf("Ошибка при шифрованиии: %v\n", err)
	}
	item.Data = encrypted

	err = storesqlite.UpdateItem(ctx, s.db, item)
	if err != nil {
		return err
	}
	return nil
}

// promptInput запрашивает у пользователя ввод текста
func promptInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// promptPassword запрашивает пароль без отображения ввода
func promptPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	// Чтение пароля без отображения символов
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

func (s *KeeperService) GetTokens(ctx context.Context) (*models.AuthResponse, error) {
	return storesqlite.GetTokens(ctx, s.db)
}
