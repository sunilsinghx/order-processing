package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	RabbitMQURL string
	RedisURL    string
	GoRoutine   string
}

func LoadEnv() {

	if os.Getenv("APP_ENV") == "docker" {
		log.Println("Running in Docker, skipping .env")
		return
	}

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println(".env not found, using existing environment variables")
	}
}

func Get(key string) string {
	return os.Getenv(key)
}

func Load() *Config {

	LoadEnv()

	return &Config{
		Port: Get("PORT"),

		DBHost:      Get("DB_HOST"),
		DBPort:      Get("DB_PORT"),
		DBUser:      Get("DB_USER"),
		DBPassword:  Get("DB_PASSWORD"),
		DBName:      Get("DB_NAME"),
		DBSSLMode:   Get("DB_SSLMODE"),
		RabbitMQURL: Get("RABBITMQ_URL"),
		RedisURL:    Get("REDIS_URL"),
		GoRoutine:   Get("GO_ROUTINE"),
	}
}
