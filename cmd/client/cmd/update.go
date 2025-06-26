package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/carinfinin/keeper/internal/service"
	"github.com/carinfinin/keeper/internal/store/models"
	"os"
	"path/filepath"

	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewUpdateCMD(cfg *clientcfg.Config) *cobra.Command {

	var uid string

	updateCMD := cobra.Command{
		Use:   "update",
		Short: "Обноваить запись по uid",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := service.NewClientService(cfg)
			if err != nil {
				pterm.Error.Printf("Ошибка инициализации: %v", err)
				os.Exit(1)
			}
			defer s.Close()

			var data []byte
			item, err := s.GetItem(cmd.Context(), uid)
			if err != nil {
				pterm.Error.Printf("Данные c uid %s не были получены: %v", uid, err)
				os.Exit(1)
			}

			data, err = processItemData(item.Type)
			if err != nil {
				pterm.Error.Printf("Ошибка обработки данных: %v", err)
				os.Exit(1)
			}

			err = s.UpdateItem(cmd.Context(), item, data)
			if err != nil {
				pterm.Error.Printf("Данные не были обновлены: %v", err)
				os.Exit(1)
			}

			pterm.DefaultSection.Printf("Запсь с uid: %s успешно обновлена", uid)
		},
	}
	updateCMD.Flags().StringVarP(&uid, "uid", "u", "", "Идентификатор записи (обязательно)")
	updateCMD.MarkFlagRequired("uid")

	return &updateCMD
}

func processItemData(itemType string) ([]byte, error) {
	switch itemType {
	case "login":
		return processLoginData()
	case "card":
		return processCardData()
	case "text":
		return processTextData()
	case "binary":
		return processBinaryData()
	default:
		return nil, fmt.Errorf("неподдерживаемый тип данных: %s", itemType)
	}
}

func processLoginData() ([]byte, error) {
	var login models.Login
	var err error

	login.Login, err = promptInput("Введите логин: ")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения логина: %w", err)
	}

	login.Password, err = promptInput("Введите пароль: ")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения пароля: %w", err)
	}

	if login.Login == "" || login.Password == "" {
		return nil, errors.New("логин и пароль не могут быть пустыми")
	}

	return json.Marshal(login)
}

func processCardData() ([]byte, error) {
	var card models.Card
	var err error

	card.Number, err = promptInput("Введите номер карты: ")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения номера карты: %w", err)
	}

	card.Expiry, err = promptInput("Введите срок действия карты: ")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения срока действия: %w", err)
	}

	card.CCV, err = promptInput("Введите CVV: ")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения CVV: %w", err)
	}

	return json.Marshal(card)
}

func processTextData() ([]byte, error) {
	text, err := promptInput("Введите текст: ")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения текста: %w", err)
	}
	return []byte(text), nil
}

func processBinaryData() ([]byte, error) {
	filePath, err := promptInput("Введите путь до файла: ")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения пути файла: %w", err)
	}

	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("файл не существует: %s", filePath)
	}

	if fileInfo.Size() > 10*1024*1024 {
		return nil, errors.New("файл слишком большой (максимум 10 МБ)")
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	fileData := models.File{
		Name:    filepath.Base(filePath),
		Size:    fileInfo.Size(),
		Content: fileContent,
	}

	return json.Marshal(fileData)
}
