package config

import (
	"time"
	"github.com/kelseyhightower/envconfig"
)

// структура содержит все необходимые конфигурации сервиса
type Config struct {
	ServerPort      string        `default:"8080" split_words:"true"`
	DatabaseURL     string        `default:"" split_words:"true"`
	StorageType     string        `default:"inmemory" split_words:"true"` // "inmemory" или "postgres"
	PostgresMaxConn int           `default:"10" split_words:"true"`
	ReadTimeout     time.Duration `default:"5s" split_words:"true"`
	WriteTimeout    time.Duration `default:"5s" split_words:"true"`
	MigrationsPath  string        `default:"pg_setup.up.sql" split_words:"true"`
}

// подгружает конфигурации из перменных окружения
func LoadConfig() (*Config, error) {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
