package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/carinfinin/keeper/internal/crypted"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/carinfinin/keeper/internal/store/storesqlite"
	"github.com/google/uuid"
	"time"

	"github.com/spf13/cobra"
)

func NewAddCMD(cfg *Config) *cobra.Command {
	var (
		typeData    string
		description string
		tags        string
		metaJSON    string
	)

	addCMD := &cobra.Command{
		Use:   "add",
		Short: "Добавление данных",
		Run: func(cmd *cobra.Command, args []string) {
			db, err := storesqlite.InitDB("test.db")
			if err != nil {
				panic(err)
			}
			defer db.Close()
			// Парсинг мета-данных
			metadata := make(map[string]string)

			salt, _ := crypted.GenerateSalt()
			pass := "1234"

			fmt.Println("metaJSON")
			fmt.Println(metaJSON)
			if metaJSON != "" {
				if err := json.Unmarshal([]byte(metaJSON), &metadata); err != nil {
					fmt.Printf("Ошибка парсинга JSON: %v\n", err)
					return
				}
			}

			var item models.Item

			item.UID = uuid.New().String()
			now := time.Now()
			item.Created = now
			item.Updated = now
			item.Meta = metadata

			switch typeData {
			case "login":
				var login models.Login

				login.Login, err = promptInput("Введите логин: ")
				if err != nil {
					fmt.Printf("Ошибка ввода: %v\n", err)
					return
				}

				login.Password, err = promptInput("Введите пароль: ")
				if err != nil {
					fmt.Printf("Ошибка ввода пароля: %v\n", err)
					return
				}

				if login.Login == "" || login.Password == "" {
					fmt.Println("Логин и пароль не могут быть пустыми")
					return
				}

				fmt.Printf("Добавлен логин: %s\nМета-данные: %v\n", login, metadata)

				data, err := json.Marshal(login)
				if err != nil {
					fmt.Println(err)
					return
				}

				encrypted, err := crypted.EncryptData(data, crypted.DeriveKey(pass, string(salt)))
				if err != nil {
					fmt.Println(err)
					return
				}

				item.Type = "login"
				item.Data = encrypted

				fmt.Println("encrypted")
				fmt.Println(encrypted)

				err = storesqlite.SaveItem(context.Background(), db, &item)
				if err != nil {
					fmt.Println(err)
				}

			default:
				fmt.Printf("Добавлены данные типа '%s'\nМета-данные: %v\n", typeData, metadata)
			}
		},
	}

	addCMD.Flags().StringVarP(&typeData, "type", "t", "text", "Тип данных (login, text, file, card)")

	// Флаги для мета-данных
	addCMD.Flags().StringVarP(&description, "desc", "d", "", "Описание данных")
	addCMD.Flags().StringVar(&tags, "tags", "", "Теги (через запятую)")
	addCMD.Flags().StringVar(&metaJSON, "meta", "", "Дополнительные мета-данные в формате JSON")

	return addCMD
}
