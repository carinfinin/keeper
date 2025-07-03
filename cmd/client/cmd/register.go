package cmd

import (
	"bufio"
	"fmt"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/service"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

// NewAuthCmd возвращает команду регистрации
func NewRegisterCmd(cfg *clientcfg.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: "Регистрация",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := service.NewClientService(cfg)
			if err != nil {
				pterm.Error.Printf("Ошибка инициализации: %v", err)
				os.Exit(1)
			}
			defer s.Close()

			var login models.Login

			login.Login, err = promptInput("Введите логин: ")
			if err != nil {
				pterm.Error.Printf("Ошибка ввода: %v\n", err)
				os.Exit(1)
			}
			login.Password, err = promptInput("Введите пароль: ")
			if err != nil {
				pterm.Error.Printf("Ошибка ввода: %v\n", err)
				os.Exit(1)
			}

			if login.Login == "" || login.Password == "" {
				pterm.Error.Println("Логин и пароль не могут быть пустыми")
				os.Exit(1)
			}

			err = s.Register(cmd.Context(), &login)
			if err != nil {
				pterm.Error.Printf("Ошибка регистрации %v\n", err)
				os.Exit(1)
			}
		},
	}
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
