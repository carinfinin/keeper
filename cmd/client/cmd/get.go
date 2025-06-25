package cmd

import (
	"github.com/carinfinin/keeper/internal/service"
	"os"

	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewGetCMD(cfg *clientcfg.Config) *cobra.Command {
	var uid string

	getCMD := cobra.Command{
		Use:   "get",
		Short: "Получить запись",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := service.NewClientService(cfg)
			if err != nil {
				pterm.Error.Printf("Ошибка инициализации: %v", err)
				os.Exit(1)
			}
			defer s.Close()

			item, err := s.GetDecryptedItem(cmd.Context(), uid)
			if err != nil {
				pterm.Error.Printf("Данные не были получены: %v", err)
				os.Exit(1)
			}

			pterm.DefaultSection.Println("Успешно получена запись:")
			pterm.Println(pterm.LightMagenta("ID:          "), pterm.LightCyan(uid))
			pterm.Println(pterm.LightMagenta("Описание:    "), pterm.LightCyan(item.Description))
			pterm.Println(pterm.LightMagenta("Данные:      "), pterm.LightCyan(string(item.Data)))
			pterm.Println(pterm.LightMagenta("Создано:     "), pterm.LightCyan(item.Created.Format("2006-01-02 15:04")))
		},
	}

	// Флаг с обязательным указанием
	getCMD.Flags().StringVarP(&uid, "uid", "u", "", "Идентификатор записи (обязательно)")
	getCMD.MarkFlagRequired("uid")

	return &getCMD
}
