package router

import "github.com/gofiber/fiber/v2"

// Setup registers all route groups.
func Setup(app *fiber.App) {
	SetupAdmin(app)
	SetupAPI(app)
}

