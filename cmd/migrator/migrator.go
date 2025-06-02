package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	var migrationsPath string

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	storageName := os.Getenv("DB_NAME")
	if storageName == "" {
		panic("DB_NAME environment variable not set")
	}

	storageAddress := os.Getenv("DB_ADDRESS")
	if storageAddress == "" {
		panic("DB_ADDRESS environment variable not set")
	}

	storageUser := os.Getenv("DB_USER")
	if storageUser == "" {
		panic("DB_USER environment variable not set")
	}

	storagePassword := os.Getenv("DB_PASSWORD")
	if storagePassword == "" {
		panic("DB_PASSWORD environment variable not set")
	}

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.Parse()

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", storageUser, storagePassword, storageAddress, storageName),
	)
	if err != nil {
		panic(err)
	}

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("migrations not found")
			return
		}

		panic(err)
	}

	fmt.Println("migrations up")
}
