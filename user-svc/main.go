package main

import (
	"github.com/SundayYogurt/user_service/config"
	"github.com/SundayYogurt/user_service/internal/api"
)

func main() {
	cfg := config.LoadConfig()

	api.StartServer(cfg)
}
