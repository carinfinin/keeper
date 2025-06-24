package clientcfg

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	BaseURL string `mapstructure:"base_url"`
	Version string `mapstructure:"version"`
	DocsURL string `mapstructure:"docs_url"`
	DBPAth  string `mapstructure:"db_path"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("base_url", "http://localhost:8080")
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("docs_url", "http://localhost:8080/docs")
	viper.SetDefault("db_path", "test.db")

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
