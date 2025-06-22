package cmd

import (
	"bufio"
	"fmt"
	"github.com/carinfinin/keeper/internal/crypted"
	"github.com/carinfinin/keeper/internal/keystore"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"net/http"
	"os"
	"strings"
	"syscall"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Регистрация",
	Run: func(cmd *cobra.Command, args []string) {

		var login models.Login

		var err error
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

		req, err := http.NewRequest(http.MethodPost, "/")

		salt, _ := crypted.GenerateSalt()
		err := keystore.SaveDerivedKey("pass", string(salt))
		if err != nil {
			fmt.Println(err)
		}
		ks, err := keystore.GetDerivedKey()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(ks.EncryptedKey)
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
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
