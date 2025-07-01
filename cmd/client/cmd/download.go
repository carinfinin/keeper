package cmd

import (
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/service"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"os"
)

// NewDownloadCMD возвращвет команду скачивания файла.
func NewDownloadCMD(cfg *clientcfg.Config) *cobra.Command {
	var (
		uid       string
		outputDir string
	)

	downloadCMD := &cobra.Command{
		Use:   "download",
		Short: "Дешифровка и скачивание файла",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := service.NewClientService(cfg)
			if err != nil {
				pterm.Error.Printf("Ошибка инициализации: %v", err)
				os.Exit(1)
			}
			defer s.Close()

			item, err := s.GetItem(cmd.Context(), uid)
			if err != nil {
				pterm.Error.Printf("Ошибка получения записи: %v", err)
				os.Exit(1)
			}

			if item.Type != "binary" {
				pterm.Error.Printf("Запись не является файлом (тип: %s)", item.Type)
				os.Exit(1)
			}

			outputPath, err := s.SaveFile(cmd.Context(), outputDir, item)

			pterm.Success.Printf("Файл успешно сохранен: %s\n", outputPath)
		},
	}

	downloadCMD.Flags().StringVarP(&uid, "uid", "u", "", "UID записи с файлом")
	downloadCMD.Flags().StringVarP(&outputDir, "output", "o", ".", "Директория для сохранения")
	downloadCMD.MarkFlagRequired("uid")

	return downloadCMD
}
