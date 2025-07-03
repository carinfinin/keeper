package cmd

import (
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/service"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"os"
)

// NewAuthCmd возвращает команду авторизации
func NewAuthCmd(cfg *clientcfg.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Авторизация",
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

			err = s.Auth(cmd.Context(), &login)
			if err != nil {
				pterm.Error.Printf("Ошибка авторизации %v\n", err)
				os.Exit(1)
			}
		},
	}
}
