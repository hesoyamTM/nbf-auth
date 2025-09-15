package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const (
	EnvConfigPath = "CONFIG_PATH"
	ConfigFlag    = "config"
)

func ParseConfigByPath[T any](path string) (*T, error) {
	const op = "config.ParseConfigByPath"

	if path == "" {
		return nil, ErrPathIsEmpty
	}

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrConfigFileNotExist
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var cfg T
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &cfg, nil
}

func MustParseConfig[T any]() *T {
	if err := godotenv.Load(); err != nil {
		log.Println("Failed to load .env file")
	}

	var cfgPath string
	flag.StringVar(&cfgPath, ConfigFlag, "", "path to config file")
	flag.Parse()

	if cfgPath == "" {
		cfgPath = os.Getenv(EnvConfigPath)
	}

	cfg, err := ParseConfigByPath[T](cfgPath)
	if err != nil {
		panic(err)
	}

	return cfg
}
