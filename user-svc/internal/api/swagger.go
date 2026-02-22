package api

import (
	_ "github.com/SundayYogurt/user_service/docs"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func RegisterSwagger(app *fiber.App) {
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
}
