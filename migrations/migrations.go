package migrations

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"graphql-comments/config"
)

func RunDatabaseMigrations(cfg *config.Config) error {
	m, err := migrate.New(
		"file://"+cfg.MigrationsPath,
		cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
