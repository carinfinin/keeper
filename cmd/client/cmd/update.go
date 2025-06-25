package cmd

import (
	"github.com/carinfinin/keeper/internal/service"
	"os"

	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewUpdateCMD(cfg *clientcfg.Config) *cobra.Command {

	var uid string

	updateCMD := cobra.Command{
		Use:   "update",
		Short: "Обноваить запись по uid",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := service.NewClientService(cfg)
			if err != nil {
				pterm.Error.Printf("Ошибка инициализации: %v", err)
				os.Exit(1)
			}
			defer s.Close()

			err = s.UpdateItem(cmd.Context(), uid)
			if err != nil {
				pterm.Error.Printf("Данные не были обновлены: %v", err)
				os.Exit(1)
			}

			pterm.DefaultSection.Printf("Запсь с uid: %s успешно обновлена", uid)
		},
	}
	updateCMD.Flags().StringVarP(&uid, "uid", "u", "", "Идентификатор записи (обязательно)")
	updateCMD.MarkFlagRequired("uid")

	return &updateCMD
}
