package service

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/store/models"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"

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

func (s *KeeperService) UpdateItem(ctx context.Context, uid string) error {

	item, err := storesqlite.GetItem(ctx, s.db, uid)
	if err != nil {
		return err
	}

	key, err := keystore.GetDerivedKey()
	if err != nil {
		return fmt.Errorf("Ошибка полученииия ключа шифрованиия: %v\n", err)
	}

	switch item.Type {
	case "login":
		var login models.Login

		login.Login, err = promptInput("Введите логин: ")
		if err != nil {
			return fmt.Errorf("Ошибка ввода: %v\n", err)
		}

		login.Password, err = promptInput("Введите пароль: ")
		if err != nil {
			return fmt.Errorf("Ошибка ввода: %v\n", err)
		}

		if login.Login == "" || login.Password == "" {
			return fmt.Errorf("Логин и пароль не могут быть пустыми")
		}

		data, err := json.Marshal(login)
		if err != nil {
			return fmt.Errorf("Ошибка перевода в json: %v\n", err)
		}

		fmt.Println(string(data))

		encrypted, err := crypted.EncryptData(data, key)
		if err != nil {
			return fmt.Errorf("Ошибка при шифрованиии: %v\n", err)
		}

		fmt.Println(string(encrypted))

		item.Data = encrypted

	default:
		fmt.Println("Добавлены данные типа ")
	}

	fmt.Println(item)

	err = storesqlite.UpdateItem(context.Background(), s.db, item)
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
