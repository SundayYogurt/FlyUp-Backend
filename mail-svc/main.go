package main

import (
	"log"

	"github.com/SundayYogurt/FlyUp-Backend/mail-svc/config"
	"github.com/SundayYogurt/FlyUp-Backend/mail-svc/infra/queue"
	"github.com/SundayYogurt/FlyUp-Backend/mail-svc/internal/api/rest/handlers"
	"github.com/SundayYogurt/FlyUp-Backend/mail-svc/internal/services"
)

func main() {
	// ---------- Load Config ----------
	cfg := config.LoadConfig()

	log.Println("Mail Service starting...")
	log.Printf("KafkaBroker=%s Topic=%s GroupID=%s\n",
		cfg.KafkaBroker,
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
	)

	// ---------- Init Service ----------
	mailService := services.NewMailService(
		cfg.GmailUser,
		cfg.GmailAppPassword,
		cfg.MailFrom,
		cfg.MailFromName,
		cfg.MailSubject,
		cfg.VerifyBaseURL,
	)

	// ---------- Init Handler ----------
	handler := handlers.NewMailHandler(mailService)

	// ---------- Init Kafka Consumer ----------
	consumer := queue.NewKafkaConsumer(
		cfg.KafkaBroker,
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
		cfg.KafkaUsername,
		cfg.KafkaPassword,
		handler,
	)

	// ---------- Start Listening ----------
	log.Println("Mail Service listening for events...")
	consumer.Listen()
}
