package cmd

import (
	"fmt"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/spf13/cobra"
)

func NewVersionCMD(cfg *clientcfg.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Версия приложения",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("v%s\n", cfg.Version)
		},
	}
}
