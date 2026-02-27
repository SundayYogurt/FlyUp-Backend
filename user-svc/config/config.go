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
	BaseURL       string
	AccessSecret  string
	KafkaUsername string
	KafkaPassword string
}

func LoadConfig() Config {
	wd, _ := os.Getwd()
	log.Println("WD =", wd)

	// เช็คว่ามีไฟล์ .env ไหม
	if _, err := os.Stat(".env"); err != nil {
		log.Println(".env not found at WD:", err)
	} else {
		log.Println(".env found")
	}

	if os.Getenv("ENV") != "prod" {
		err := godotenv.Overload()
		if err != nil {
			log.Println("Warning: env file not found or could not be loaded:", err)
		}
	}

	return Config{
		ServerPort:    os.Getenv("SERVER_PORT"),
		DatabaseDSN:   os.Getenv("DATABASE_DSN"),
		KafkaBroker:   os.Getenv("KAFKA_BROKER"),
		KafkaTopic:    os.Getenv("KAFKA_TOPIC"),
		CloudinaryUrl: os.Getenv("CLOUDINARY_URL"),
		IAppApiKey:    os.Getenv("IAPP_API_KEY"),
		BaseURL:       os.Getenv("BASE_URL"),
		AccessSecret:  os.Getenv("ACCESS_SECRET"),
		KafkaUsername: os.Getenv("KAFKA_USERNAME"),
		KafkaPassword: os.Getenv("KAFKA_PASSWORD"),
	}
}
