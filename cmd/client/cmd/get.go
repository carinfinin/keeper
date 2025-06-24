package cmd

import (
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/crypted"
	"github.com/carinfinin/keeper/internal/keystore"
	"github.com/carinfinin/keeper/internal/store/storesqlite"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewGetCMD(cfg *clientcfg.Config) *cobra.Command {

	var uid string

	getCMD := cobra.Command{
		Use:   "get",
		Short: "Получить запись",
		Run: func(cmd *cobra.Command, args []string) {

			key, err := keystore.GetDerivedKey()
			if err != nil {
				pterm.Error.Printf("Ключ шифрования не был получен error: %v\n", err)
				return
			}

			db, err := storesqlite.InitDB(cfg.DBPAth)
			if err != nil {
				pterm.Error.Printf("бд не запущена error: %v\n", err)
				return
			}

			item, err := storesqlite.GetItem(cmd.Context(), db, uid)
			if err != nil {
				pterm.Error.Printf("запись не была получена error: %v\n", err)
				return
			}

			decrypt, err := crypted.DecryptData(item.Data, key)
			if err != nil {
				pterm.Error.Printf("запись не была засшифрована error: %v\n", err)
				return
			}

			pterm.Info.Printf("%s, %s\n", decrypt, item.Description)
		},
	}
	getCMD.Flags().StringVar(&uid, "uid", "", "Идентификатор записи")

	return &getCMD
}
