package cmd

import (
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/service"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// NewAddCMD возвращает команду добавления данных
func NewAddCMD(cfg *clientcfg.Config) *cobra.Command {
	var (
		typeData    string
		description string
	)

	addCMD := &cobra.Command{
		Use:   "add",
		Short: "Добавление данных",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := service.NewClientService(cfg)
			if err != nil {
				pterm.Error.Printf("Ошибка инициализации: %v", err)
				os.Exit(1)
			}
			defer s.Close()

			var item models.Item
			var data []byte

			item.UID = uuid.New().String()
			now := time.Now()
			item.Created = now
			item.Updated = now
			item.Type = typeData
			item.Description = description

			data, err = processItemData(item.Type)
			if err != nil {
				pterm.Error.Printf("Ошибка обработки данных: %v", err)
				os.Exit(1)
			}
			err = s.AddDecryptedItem(cmd.Context(), &item, data)
			if err != nil {
				pterm.Error.Printf("Данные не были добаввлены: %v", err)
				os.Exit(1)
			}

			pterm.DefaultSection.Printf("Запсь успешно c uid: %s добавлена\n", item.UID)

		},
	}

	addCMD.Flags().StringVarP(&typeData, "type", "t", "text", "Тип данных (login, text, file, card)")
	addCMD.Flags().StringVarP(&description, "desc", "d", "", "Описание данных")

	return addCMD
}
