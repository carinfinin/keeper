package cmd

import (
	"fmt"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/spf13/cobra"
	"os"
)

func Execute() {

	cfg, err := clientcfg.LoadConfig()
	if err != nil {
		fmt.Printf("Ошибка загрузки конфига: %v\n", err)
		os.Exit(1)
	}

	cmd := NewRootCommand(cfg)

	if err = cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewRootCommand(cfg *clientcfg.Config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "app",
		Short: "Мое приложение",
	}
	rootCmd.AddCommand(
		NewRegisterCmd(cfg),
		NewVersionCMD(cfg),
		NewAddCMD(cfg),
		NewAuthCmd(cfg),
		NewGetCMD(cfg),
		NewListCMD(cfg),
		NewDeleteCMD(cfg),
		NewUpdateCMD(cfg),
	)
	return rootCmd
}
