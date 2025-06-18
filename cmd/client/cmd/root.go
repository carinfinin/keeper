package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var name string
var uppercase bool

// rootCmd - базовая команда
var rootCmd = &cobra.Command{
	Use:   "greet",
	Short: "Приветствие пользователя",
	Run: func(cmd *cobra.Command, args []string) {
		message := fmt.Sprintf("Привет, %s!", name)
		if uppercase {
			message = strings.ToUpper(message)
		}
		fmt.Println(message)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&name, "name", "n", "World", "Ваше имя")
	rootCmd.Flags().BoolVarP(&uppercase, "uppercase", "u", false, "Верхний регистр")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
