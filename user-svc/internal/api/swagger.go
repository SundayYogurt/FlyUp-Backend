package api

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/gofiber/swagger"

	// IMPORTANT: เปลี่ยน path ให้ตรงกับ module ของคุณ
	// ตัวอย่าง: module = github.com/SundayYogurt/user_service
	"github.com/SundayYogurt/user_service/docs"
)

type SwaggerConfig struct {
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
	Schemes     []string // []string{"http","https"}
}

func RegisterSwagger(app *fiber.App, cfg SwaggerConfig) {
	// set swagger info
	docs.SwaggerInfo.Title = cfg.Title
	docs.SwaggerInfo.Description = cfg.Description
	docs.SwaggerInfo.Version = cfg.Version
	docs.SwaggerInfo.Host = cfg.Host
	docs.SwaggerInfo.BasePath = cfg.BasePath
	if len(cfg.Schemes) > 0 {
		docs.SwaggerInfo.Schemes = cfg.Schemes
	}

	// route: /swagger/index.html
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)
}
