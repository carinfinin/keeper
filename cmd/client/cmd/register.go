package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/fingerprint"
	"github.com/carinfinin/keeper/internal/keystore"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/carinfinin/keeper/internal/store/storesqlite"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
)

func NewRegisterCmd(cfg *clientcfg.Config) *cobra.Command {
	return &cobra.Command{
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

			bl, err := json.Marshal(login)
			if err != nil {
				fmt.Println(err)
				return
			}

			req, err := http.NewRequest(http.MethodPost, cfg.BaseURL+"/api/register", bytes.NewReader(bl))
			if err != nil {
				fmt.Println(err)
				return
			}

			fp := fingerprint.Get()
			deviceID := fp.GenerateHash()

			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("User-Agent", fmt.Sprintf("Kepper/%s (*%s)", cfg.Version, deviceID))

			client := http.DefaultClient

			var ar models.AuthResponse
			response, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}

			if response.StatusCode != http.StatusCreated {
				fmt.Println("error status code register: ", response.Status)
				return
			}

			b, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer response.Body.Close()

			err = json.Unmarshal(b, &ar)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = saveCredentials(cmd.Context(), cfg, &ar)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(ar)

			err = keystore.SaveDerivedKey(login.Password, ar.Salt)
			if err != nil {
				fmt.Println(err)
			}
			//ks, err := keystore.GetDerivedKey()
			//if err != nil {
			//	fmt.Println(err)
			//}
			//
			//fmt.Println(ks.EncryptedKey)
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

func saveCredentials(ctx context.Context, сfg *clientcfg.Config, a *models.AuthResponse) error {

	db, err := storesqlite.InitDB(сfg.DBPAth)
	if err != nil {
		return err
	}
	defer db.Close()
	return storesqlite.UpsertTokens(ctx, db, a)
}
