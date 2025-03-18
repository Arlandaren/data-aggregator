package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"service/internal/infrastructure/config"
	"time"
)

func main() {
	var migrationsPath, migrationsTable string
	var down, rollback, force bool
	var version int

	cfg := config.NewConfig()

	pgString := cfg.Postgres.ConnStr

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.BoolVar(&down, "down", false, "set this flag to revert all migrations")
	flag.BoolVar(&rollback, "rollback", false, "set this flag to rollback the last migration")
	flag.BoolVar(&force, "force", false, "set this flag to force a specific migration version")
	flag.IntVar(&version, "version", 0, "migration version to force (used with --force)")
	flag.Parse()

	if migrationsPath == "" {
		log.Fatal("migrations-path is required")
	}

	pgStringWithTable := fmt.Sprintf("%s&x-migrations-table=%s", pgString, migrationsTable)

	for {
		m, err := migrate.New(
			"file://"+migrationsPath,
			pgStringWithTable,
		)
		if err != nil {
			log.Printf("failed to initialize migrate instance: %v. Retrying in 10 seconds...", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if force {
			if version == 0 {
				log.Fatal("version is required with --force flag")
			}
			if err := m.Force(version); err != nil {
				log.Printf("failed to force migration version: %v. Retrying in 10 seconds...", err)
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Printf("forced migration version to %d\n", version)
		} else if down {
			if err := m.Down(); err != nil {
				if errors.Is(err, migrate.ErrNoChange) {
					fmt.Println("no migrations to revert")
					return
				}
				log.Printf("failed to revert all migrations: %v. Retrying in 10 seconds...", err)
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Println("all migrations reverted successfully")
		} else if rollback {
			if err := m.Steps(-1); err != nil {
				if errors.Is(err, migrate.ErrNoChange) {
					fmt.Println("no migration to rollback")
					return
				}
				log.Printf("failed to rollback migration: %v. Retrying in 10 seconds...", err)
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Println("last migration rolled back successfully")
		} else {
			if err := m.Up(); err != nil {
				if errors.Is(err, migrate.ErrNoChange) {
					fmt.Println("no migrations to apply")
					return
				}
				log.Printf("failed to apply migrations: %v. Retrying in 10 seconds...", err)
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Println("migrations applied successfully")
		}
		// Если миграция успешна, выходим из цикла
		break
	}
}

type Log struct {
	verbose bool
}

func (l *Log) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (l *Log) Verbose() bool {
	return l.verbose
}
