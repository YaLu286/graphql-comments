package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Очищаем переменные окружения, функция будет использовать дефолтные значения
	os.Clearenv()

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "8080", config.ServerPort)
	assert.Equal(t, "", config.DatabaseURL)
	assert.Equal(t, "inmemory", config.StorageType)
	assert.Equal(t, 10, config.PostgresMaxConn)
	assert.Equal(t, 5*time.Second, config.ReadTimeout)
	assert.Equal(t, 5*time.Second, config.WriteTimeout)
	assert.Equal(t, "pg_setup.up.sql", config.MigrationsPath)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DATABASE_URL", "postgres://user:password@localhost/postgres")
	os.Setenv("STORAGE_TYPE", "postgres")
	os.Setenv("POSTGRES_MAX_CONN", "20")
	os.Setenv("READ_TIMEOUT", "10s")
	os.Setenv("WRITE_TIMEOUT", "10s")
	os.Setenv("MIGRATIONS_PATH", "migrations.sql")

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "9090", config.ServerPort)
	assert.Equal(t, "postgres://user:password@localhost/postgres", config.DatabaseURL)
	assert.Equal(t, "postgres", config.StorageType)
	assert.Equal(t, 20, config.PostgresMaxConn)
	assert.Equal(t, 10*time.Second, config.ReadTimeout)
	assert.Equal(t, 10*time.Second, config.WriteTimeout)
	assert.Equal(t, "migrations.sql", config.MigrationsPath)
}
