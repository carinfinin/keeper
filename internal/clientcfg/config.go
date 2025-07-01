package clientcfg

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// Config для клиента.
type Config struct {
	BaseURL string `mapstructure:"base_url"`
	Version string `mapstructure:"version"`
	DocsURL string `mapstructure:"docs_url"`
	DBPAth  string `mapstructure:"db_path"`
}

// LoadConfig конструктор.
func LoadConfig() (*Config, error) {
	viper.SetDefault("base_url", "http://localhost:8080")
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("docs_url", "http://localhost:8080/docs")
	viper.SetDefault("db_path", "keeper.db")

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	viper.AddConfigPath(filepath.Join(home, ".keeper"))
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err = viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
