package main

import (
	"log"

	"github.com/SundayYogurt/user_service/config"
	"github.com/SundayYogurt/user_service/infra/queue"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Content-Type , Accept, Authorization",
	}))

	//load configuration
	cfg := config.LoadConfig()
	kafkaProducer := queue.NewProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	log.Printf("Kafka producer created: %v", kafkaProducer)

	app.Get("/", HealthCheck)

	app.Listen("localhost:3000")

}

func HealthCheck(ctx *fiber.Ctx) error {
	return ctx.Status(200).JSON(fiber.Map{
		"message": "Healthy!",
	})
}
