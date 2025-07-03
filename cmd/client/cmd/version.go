package cmd

import (
	"fmt"
	"github.com/carinfinin/keeper/internal/buildinfo"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/spf13/cobra"
)

// NewVersionCMD возвращвет команду версии сборки.
func NewVersionCMD(cfg *clientcfg.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Версия приложения",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Версия приложения: %s\n", buildinfo.Version)
			fmt.Printf("Дата сборки: %s\n", buildinfo.BuildDate)
		},
	}
}
