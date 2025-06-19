package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/carinfinin/keeper/internal/store/storesqlite"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	typeData    string
	description string
	tags        string
	metaJSON    string
)

var addCmd = &cobra.Command{
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

		if metaJSON != "" {
			if err := json.Unmarshal([]byte(metaJSON), &metadata); err != nil {
				fmt.Printf("Ошибка парсинга JSON: %v\n", err)
				return
			}
		}

		// Обработка в зависимости от типа
		switch typeData {
		case "login":
			// Интерактивный ввод логина и пароля
			login, err := promptInput("Введите логин: ")
			if err != nil {
				fmt.Printf("Ошибка ввода: %v\n", err)
				return
			}

			password, err := promptPassword("Введите пароль: ")
			if err != nil {
				fmt.Printf("Ошибка ввода пароля: %v\n", err)
				return
			}

			// Проверка введенных данных
			if login == "" || password == "" {
				fmt.Println("Логин и пароль не могут быть пустыми")
				return
			}

			fmt.Printf("Добавлен логин: %s\nМета-данные: %v\n", login, metadata)
			// Здесь можно сохранить данные, например, в зашифрованном виде

		default:
			fmt.Printf("Добавлены данные типа '%s'\nМета-данные: %v\n", typeData, metadata)
		}
	},
}

func init() {
	// Основные флаги
	addCmd.Flags().StringVarP(&typeData, "type", "t", "text", "Тип данных (login, text, file, card)")

	// Флаги для мета-данных
	addCmd.Flags().StringVarP(&description, "desc", "d", "", "Описание данных")
	addCmd.Flags().StringVar(&tags, "tags", "", "Теги (через запятую)")
	addCmd.Flags().StringVar(&metaJSON, "meta", "", "Дополнительные мета-данные в формате JSON")

	rootCmd.AddCommand(addCmd)
}

// Вспомогательные функции для интерактивного ввода

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
