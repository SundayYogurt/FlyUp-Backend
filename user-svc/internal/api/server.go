// @title FlyUp User Service API
// @version 1.0
// @description Auth/Profile/KYC/Pioneer/Admin endpoints.
// @host localhost:3000
// @BasePath /
// @schemes http
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer <JWT>

package api

import (
	"log"

	"github.com/SundayYogurt/user_service/config"
	"github.com/SundayYogurt/user_service/infra/queue"
	"github.com/SundayYogurt/user_service/internal/api/rest/handlers"
	"github.com/SundayYogurt/user_service/internal/clients/iapp"
	"github.com/SundayYogurt/user_service/internal/domain"
	"github.com/SundayYogurt/user_service/internal/helper"
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
	RegisterSwagger(app)
	log.Printf("KafkaBroker=%q KafkaTopic=%q", cfg.KafkaBroker, cfg.KafkaTopic)
	// ---------- CORS ----------
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.BaseURL,
		AllowHeaders:     "Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// ---------- DB ----------
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DatabaseDSN,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	log.Println("database connected")

	// ---------- MIGRATION + SEED (guarded by advisory lock) ----------
	// ใช้เลขคงที่ตัวเดียวกันทั้งระบบเพื่อ lock งาน migrate
	const migrateLockID int64 = 20260222

	if err := db.Exec("SELECT pg_advisory_lock(?)", migrateLockID).Error; err != nil {
		log.Fatalf("migration lock error: %v", err)
	}
	defer func() {
		_ = db.Exec("SELECT pg_advisory_unlock(?)", migrateLockID).Error
	}()

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Role{},
		&domain.UserRole{},
		&domain.University{},
		&domain.StudentProfile{},
		&domain.KYCSubmission{},
		&domain.KYCDocument{},
		&domain.KYCReview{},
		&domain.UserConsent{},
	); err != nil {
		log.Fatalf("migration error: %v", err)
	}
	log.Println("migration successful")

	seedRoles(db)

	// ---------- Infra ----------
	kafkaProducer := queue.NewProducer(
		cfg.KafkaBroker,
		cfg.KafkaTopic,
		cfg.KafkaUsername,
		cfg.KafkaPassword,
	)
	cld, err := cloudinary.New()
	if err != nil {
		log.Fatalf("cloudinary init error: %v", err)
	}
	iappClient := iapp.New(cfg.IAppApiKey)
	up := cloudinary.NewCloudinaryUploader(cld)

	authHelper := helper.SetupAuth(cfg.AccessSecret)

	// ---------- Repositories ----------
	userRepo := repository.NewUserRepository(db)
	kycRepo := repository.NewKYCRepository(db)
	studentRepo := repository.NewStudentProfileRepository(db)
	universityRepo := repository.NewUniversityRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	bankAccountRepo := repository.NewBankAccountRepository(db)
	consentRepo := repository.NewConsentRepository(db)
	// ---------- Service ----------
	userSvc := services.NewUserService(
		userRepo,
		kafkaProducer,
		kycRepo,
		studentRepo,
		universityRepo,
		roleRepo,
		userRoleRepo,
		bankAccountRepo,
		iappClient,
		up,
		consentRepo,
		authHelper,
	)

	// ---------- Handler ----------
	userHandler := handlers.NewUserHandler(userSvc, cld, authHelper)
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
