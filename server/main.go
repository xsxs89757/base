package main

import (
	"fmt"
	"log"

	"base/config"
	_ "base/docs"
	"base/internal/middleware"
	"base/internal/router"
	"base/internal/store"
	"base/internal/validator"
	_ "base/internal/validator/admin"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

// @title Admin 后台管理系统 API
// @version 1.0
// @description 基于 Go Fiber + GORM + Casbin + JWT 的后台管理系统 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email admin@example.com

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 Bearer {token} 格式的 JWT 令牌

func main() {
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	validator.Init()
	store.Init()
	middleware.InitCasbin()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"code":    -1,
				"data":    nil,
				"error":   err.Error(),
				"message": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5666,http://localhost:5173,http://localhost:8080,http://127.0.0.1:5666",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,Accept-Language",
		AllowCredentials: true,
	}))

	// Swagger UI (仅开发环境)
	if config.C.Server.EnableSwagger {
		app.Get("/swagger/*", swagger.HandlerDefault)
		log.Printf("Swagger UI: http://localhost:%d/swagger/index.html", config.C.Server.Port)
		log.Printf("OpenAPI JSON: http://localhost:%d/swagger/doc.json", config.C.Server.Port)
	}

	router.Setup(app)

	addr := fmt.Sprintf(":%d", config.C.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Fatal(app.Listen(addr))
}
