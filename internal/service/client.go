package service

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/fingerprint"
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

// KeeperService представляет основной сервис приложения, объединяющий:
// - Работу с локальной базой данных
// - Сетевое взаимодействие с сервером
// - Управление аутентификацией
type KeeperService struct {
	db  *sql.DB
	cfg *clientcfg.Config
}

// NewClientService создает новый экземпляр KeeperService.
func NewClientService(cfg *clientcfg.Config) (*KeeperService, error) {
	db, err := storesqlite.InitDB(cfg.DBPAth)
	if err != nil {
		return nil, err
	}
	return &KeeperService{db: db, cfg: cfg}, nil
}

// Close.
func (s *KeeperService) Close() error {
	return s.db.Close()
}

// GetDecryptedItem возвращает расшифрованную запись по её UID.
// Автоматически дешифрует данные используя мастер-ключ.
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

// GetItem возвращает зашифрованную запись по её UID без дешифровки.
func (s *KeeperService) GetItem(ctx context.Context, uid string) (*models.Item, error) {
	return storesqlite.GetItem(ctx, s.db, uid)
}

// RefreshTokens обновляет пару access/refresh токенов.
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

// SaveFile сохраняет зашифрованный файл из хранилища на диск.
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

// AddDecryptedItem добавляет новую запись с предварительно зашифрованными данными.
// Шифрует переданные данные перед сохранением в БД.
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

// GetDecryptedItems возвращает все записи с расшифрованным содержимым.
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

// DeleteItem удаляет запись UID.
func (s *KeeperService) DeleteItem(ctx context.Context, uid string) error {
	return storesqlite.DeleteItem(ctx, s.db, uid)
}

// GetLastChanges возвращает записи, измененные после указанной даты.
func (s *KeeperService) GetLastChanges(ctx context.Context, lastSync time.Time) ([]*models.Item, error) {
	return storesqlite.GetLastItems(ctx, s.db, lastSync)
}

// MergeLastChanges применяет изменения из переданного списка записей.
func (s *KeeperService) MergeLastChanges(ctx context.Context, items []*models.Item) error {
	return storesqlite.UpdateItems(ctx, s.db, items)
}

// UpdateItem обновляет существующую запись.
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

// GetTokens возвращает текущие аутентификационные токены из локальной БД.
func (s *KeeperService) GetTokens(ctx context.Context) (*models.AuthResponse, error) {
	return storesqlite.GetTokens(ctx, s.db)
}

// saveCredentials сохраняет токены в локальную БД.
func (s *KeeperService) saveCredentials(ctx context.Context, a *models.AuthResponse) error {
	return storesqlite.UpsertTokens(ctx, s.db, a)
}

// Auth выполняет аутентификацию на сервере.
func (s *KeeperService) Auth(ctx context.Context, login *models.Login) error {
	bl, err := json.Marshal(login)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.cfg.BaseURL+"/api/login", bytes.NewReader(bl))
	if err != nil {
		return err
	}

	fp := fingerprint.Get()
	deviceID := fp.GenerateHash()

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("Kepper/%s (*%s)", s.cfg.Version, deviceID))

	client := http.DefaultClient

	var ar models.AuthResponse
	response, err := client.Do(req)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error status code auth: ", response.Status)
	}

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.Unmarshal(b, &ar)
	if err != nil {
		return err
	}

	err = s.saveCredentials(ctx, &ar)
	if err != nil {
		return err
	}
	return keystore.SaveDerivedKey(login.Password, ar.Salt)
}

// Register регистрирует нового пользователя на сервере.
func (s *KeeperService) Register(ctx context.Context, login *models.Login) error {
	bl, err := json.Marshal(login)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.cfg.BaseURL+"/api/register", bytes.NewReader(bl))
	if err != nil {
		return err
	}

	fp := fingerprint.Get()
	deviceID := fp.GenerateHash()

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("Kepper/%s (*%s)", s.cfg.Version, deviceID))

	client := http.DefaultClient

	var ar models.AuthResponse
	response, err := client.Do(req)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("error status code auth: ", response.Status)
	}

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.Unmarshal(b, &ar)
	if err != nil {
		return err
	}

	err = s.saveCredentials(ctx, &ar)
	if err != nil {
		return err
	}
	return keystore.SaveDerivedKey(login.Password, ar.Salt)
}
