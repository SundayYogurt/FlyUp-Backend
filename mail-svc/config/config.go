package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBroker      string
	KafkaTopic       string
	KafkaGroupID     string
	MailProvider     string
	GmailUser        string
	GmailAppPassword string
	MailFrom         string
	MailFromName     string
	VerifyBaseURL    string
	MailSubject      string
	KafkaUsername    string
	KafkaPassword    string
}

func LoadConfig() Config {
	wd, _ := os.Getwd()
	log.Println("WD =", wd)

	if os.Getenv("ENV") != "prod" {
		if err := godotenv.Overload(); err != nil {
			log.Println("Warning: .env not loaded:", err)
		}
	}

	return Config{
		KafkaBroker:      os.Getenv("KAFKA_BROKER"),
		KafkaTopic:       os.Getenv("KAFKA_TOPIC"),
		KafkaGroupID:     os.Getenv("KAFKA_GROUP_ID"),
		MailProvider:     os.Getenv("MAIL_PROVIDER"),
		GmailUser:        os.Getenv("GMAIL_USER"),
		GmailAppPassword: os.Getenv("GMAIL_APP_PASSWORD"),
		MailFrom:         os.Getenv("MAIL_FROM"),
		MailFromName:     os.Getenv("MAIL_FROM_NAME"),
		VerifyBaseURL:    os.Getenv("VERIFY_BASE_URL"),
		MailSubject:      os.Getenv("MAIL_SUBJECT"),
		KafkaUsername:    os.Getenv("KAFKA_USERNAME"),
		KafkaPassword:    os.Getenv("KAFKA_PASSWORD"),
	}
}
