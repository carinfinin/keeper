package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/crypted"
	"github.com/carinfinin/keeper/internal/keystore"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/carinfinin/keeper/internal/store/storesqlite"
	"github.com/google/uuid"
	"time"

	"github.com/spf13/cobra"
)

func NewAddCMD(cfg *clientcfg.Config) *cobra.Command {
	var (
		typeData    string
		description string
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

			//salt, _ := crypted.GenerateSalt()
			//pass := "1234"

			var item models.Item

			item.UID = uuid.New().String()
			now := time.Now()
			item.Created = now
			item.Updated = now
			item.Description = description

			key, err := keystore.GetDerivedKey()
			if err != nil {
				fmt.Println(err)
				return
			}

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

				fmt.Printf("Добавлен логин: %s\nМета-данные: %v\n", login, description)

				data, err := json.Marshal(login)
				if err != nil {
					fmt.Println(err)
					return
				}

				encrypted, err := crypted.EncryptData(data, key)
				if err != nil {
					fmt.Println(err)
					return
				}

				item.Type = "login"
				item.Data = encrypted

				fmt.Println("encrypted")
				fmt.Println(encrypted)

			default:
				fmt.Printf("Добавлены данные типа '%s'\nМета-данные: %v\n", typeData, description)
			}

			err = storesqlite.SaveItem(context.Background(), db, &item)
			if err != nil {
				fmt.Println(err)
			}
		},
	}

	addCMD.Flags().StringVarP(&typeData, "type", "t", "text", "Тип данных (login, text, file, card)")

	// Флаги для мета-данных
	addCMD.Flags().StringVarP(&description, "desc", "d", "", "Описание данных")

	return addCMD
}
