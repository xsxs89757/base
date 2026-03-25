package router

import "github.com/gofiber/fiber/v2"

// SetupAPI registers all /api/* routes (public-facing API).
// Currently a placeholder — add public endpoints here as needed.
func SetupAPI(app *fiber.App) {
	_ = app.Group("/api")
}
