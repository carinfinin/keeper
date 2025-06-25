package cmd

import (
	"github.com/carinfinin/keeper/internal/service"
	"os"

	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewListCMD(cfg *clientcfg.Config) *cobra.Command {

	return &cobra.Command{
		Use:   "list",
		Short: "Получить все записи",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := service.NewClientService(cfg)
			if err != nil {
				pterm.Error.Printf("Ошибка инициализации: %v", err)
				os.Exit(1)
			}
			defer s.Close()

			items, err := s.GetDecryptedItems(cmd.Context())
			if err != nil {
				pterm.Error.Printf("Данные не были получены: %v", err)
				os.Exit(1)
			}

			if len(items) == 0 {
				pterm.Error.Println("У вас нет сохранённых данных")
				os.Exit(1)
			}

			pterm.DefaultSection.Println("Успешно получены записи:")
			for _, item := range items {

				pterm.Println(pterm.LightMagenta("ID:          "), pterm.LightCyan(item.UID))
				pterm.Println(pterm.LightMagenta("Описание:    "), pterm.LightCyan(item.Description))
				pterm.Println(pterm.LightMagenta("Данные:      "), pterm.LightCyan(string(item.Data)))
				pterm.Println(pterm.LightMagenta("Создано:     "), pterm.LightCyan(item.Created.Format("2006-01-02 15:04")))
				pterm.Println()
			}

		},
	}
}
