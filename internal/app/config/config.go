package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

var Conf Config

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DataBaseDSN     string `end:"DATABASE_DSN"`
}

var f flagConfig

type flagConfig struct {
	ServerAddress   *string
	BaseURL         *string
	FileStoragePath *string
	DataBaseDSN     *string
}

func init() {
	f.ServerAddress = flag.String("a", "localhost:8080", "server address")
	f.BaseURL = flag.String("b", "", "base url")
	f.FileStoragePath = flag.String("f", "", "file storage path")
	f.DataBaseDSN = flag.String("d", "", "database address")
}

func ParseConfig() (Config, error) {
	flag.Parse()
	Conf.ServerAddress = *f.ServerAddress
	Conf.FileStoragePath = *f.FileStoragePath
	Conf.BaseURL = *f.BaseURL
	Conf.DataBaseDSN = *f.DataBaseDSN

	err := env.Parse(&Conf)
	if err != nil {
		return Config{}, err
	}

	if Conf.ServerAddress == "" {
		Conf.ServerAddress = "localhost:8080"
	}

	if Conf.ServerAddress != "" {
		Conf.ServerAddress += "/"
	}

	if Conf.BaseURL != "" {
		Conf.BaseURL += "/"
	}

	return Conf, nil
}
