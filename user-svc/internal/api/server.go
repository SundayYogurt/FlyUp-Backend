package api

import (
	"log"

	"github.com/SundayYogurt/user_service/config"
	"github.com/SundayYogurt/user_service/infra/queue"
	"github.com/SundayYogurt/user_service/internal/api/rest/handlers"
	"github.com/SundayYogurt/user_service/internal/clients/iapp"
	"github.com/SundayYogurt/user_service/internal/domain"
	"github.com/SundayYogurt/user_service/internal/repository"
	"github.com/SundayYogurt/user_service/internal/services"
	"github.com/SundayYogurt/user_service/pkg/cloudinary"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func StartServer(cfg config.Config) {
	app := fiber.New()

	// ---------- CORS ----------
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	// ---------- DB ----------
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	log.Println("database connected")

	seedRoles(db)

	// ---------- MIGRATION ----------
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Role{},
		&domain.UserRole{},
		&domain.University{},
		&domain.StudentProfile{},
		&domain.KYCSubmission{},
		&domain.KYCDocument{},
		&domain.KYCReview{},
	); err != nil {
		log.Fatalf("migration error: %v", err)
	}
	log.Println("migration successful")

	// ---------- Infra ----------
	kafkaProducer := queue.NewProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	cld, err := cloudinary.New()
	if err != nil {
		log.Fatalf("cloudinary init error: %v", err)
	}
	iappClient := iapp.New(cfg.IAppApiKey)
	up := cloudinary.NewCloudinaryUploader(cld)

	// ---------- Repositories ----------
	userRepo := repository.NewUserRepository(db)
	kycRepo := repository.NewKYCRepository(db)
	studentRepo := repository.NewStudentProfileRepository(db)
	universityRepo := repository.NewUniversityRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)

	// ---------- Service ----------
	userSvc := services.NewUserService(
		userRepo,
		kafkaProducer,
		kycRepo,
		studentRepo,
		universityRepo,
		roleRepo,
		userRoleRepo,
		iappClient,
		up,
	)

	// ---------- Handler ----------
	userHandler := handlers.NewUserHandler(
		userSvc,
		cld,
	)
	userHandler.SetupRoutes(app)

	// ---------- Health ----------
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ---------- Listen ----------
	addr := cfg.ServerPort
	log.Println("listening on", addr)
	log.Fatal(app.Listen(addr))
}

func seedRoles(db *gorm.DB) {
	codes := []string{"ADMIN", "BOOSTER", "PIONEER"}

	for _, code := range codes {
		var r domain.Role
		err := db.Where("code = ?", code).First(&r).Error
		if err == gorm.ErrRecordNotFound {
			_ = db.Create(&domain.Role{
				Code: code,
				Name: code, //
			}).Error
		}
	}
}
