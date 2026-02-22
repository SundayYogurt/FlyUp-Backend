package api

import (
	docs "github.com/SundayYogurt/user_service/docs" // ใช้ set Host/Schemes
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func RegisterSwagger(app *fiber.App) {
	// set ตอน runtime ให้ตรงโดเมน/โปรโตคอลที่ user เข้า (http หรือ https)
	app.Use(func(c *fiber.Ctx) error {
		// Host เช่น "flyup-user-svc.onrender.com"
		docs.SwaggerInfo.Host = c.Hostname()

		// Protocol: "http" หรือ "https"
		docs.SwaggerInfo.Schemes = []string{c.Protocol()}
		return c.Next()
	})

	app.Get("/swagger/*", fiberSwagger.WrapHandler)
}
