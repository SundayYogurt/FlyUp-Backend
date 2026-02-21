package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	DatabaseDSN   string
	KafkaBroker   string
	KafkaTopic    string
	CloudinaryUrl string
	IAppApiKey    string
}

func LoadConfig() Config {
	if os.Getenv("ENV") != "prod" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: env file not found or could not be loaded")
		}
	}

	return Config{
		ServerPort:    os.Getenv("SERVER_PORT"),
		DatabaseDSN:   os.Getenv("DATABASE_DSN"),
		KafkaBroker:   os.Getenv("KAFKA_BROKER"),
		KafkaTopic:    os.Getenv("KAFKA_TOPIC"),
		CloudinaryUrl: os.Getenv("CLOUDINARY_URL"),
		IAppApiKey:    os.Getenv("IAPP_API_KEY"),
	}
}
