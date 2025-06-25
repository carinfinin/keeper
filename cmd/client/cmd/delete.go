package cmd

import (
	"github.com/carinfinin/keeper/internal/service"
	"os"

	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewDeleteCMD(cfg *clientcfg.Config) *cobra.Command {
	var uid string

	deleteCMD := cobra.Command{
		Use:   "delete",
		Short: "Удалить запись по uid",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := service.NewClientService(cfg)
			if err != nil {
				pterm.Error.Printf("Ошибка инициализации: %v", err)
				os.Exit(1)
			}
			defer s.Close()

			err = s.DeleteItem(cmd.Context(), uid)
			if err != nil {
				pterm.Error.Printf("Данные не были удалены: %v", err)
				os.Exit(1)
			}

			pterm.DefaultSection.Printf("Запсь с uid: %s успешно удалена", uid)
		},
	}
	deleteCMD.Flags().StringVarP(&uid, "uid", "u", "", "Идентификатор записи (обязательно)")
	deleteCMD.MarkFlagRequired("uid")

	return &deleteCMD
}
