package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"time"
)

type Config struct {
	Env        string `env:"ENV" env-default:"local"`
	Storage    Storage
	HTTPServer HTTPServer
}

type Storage struct {
	Address  string `env:"DB_ADDRESS" env-required:"true"`
	User     string `env:"DB_USER" env-required:"true"`
	Name     string `env:"DB_NAME" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `env:"SERVER_ADDRESS" env-default:"localhost:8081"`
	Timeout     time.Duration `env:"SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	var config Config

	if err := cleanenv.ReadEnv(&config); err != nil {
		log.Fatal("cannot read config: ", err)
	}

	return &config
}
