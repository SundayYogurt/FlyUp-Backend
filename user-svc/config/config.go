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
	wd, _ := os.Getwd()
	log.Println("WD =", wd)

	// เช็คว่ามีไฟล์ .env ไหม
	if _, err := os.Stat(".env"); err != nil {
		log.Println(".env not found at WD:", err)
	} else {
		log.Println(".env found")
	}

	if os.Getenv("ENV") != "prod" {
		err := godotenv.Overload() // ใช้ Overload ให้ชัวร์
		if err != nil {
			log.Println("Warning: env file not found or could not be loaded:", err)
		}
	}

	log.Println("IAPP_API_KEY =", os.Getenv("IAPP_API_KEY"))

	return Config{
		ServerPort:    os.Getenv("SERVER_PORT"),
		DatabaseDSN:   os.Getenv("DATABASE_DSN"),
		KafkaBroker:   os.Getenv("KAFKA_BROKER"),
		KafkaTopic:    os.Getenv("KAFKA_TOPIC"),
		CloudinaryUrl: os.Getenv("CLOUDINARY_URL"),
		IAppApiKey:    os.Getenv("IAPP_API_KEY"),
	}
}
