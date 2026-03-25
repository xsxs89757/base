package middleware

import (
	"strings"
	"time"

	adminmodel "base/internal/model/admin"
	"base/internal/store"

	"github.com/gofiber/fiber/v2"
)

func OperationLog() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "GET" || c.Method() == "OPTIONS" {
			return c.Next()
		}

		path := c.Path()
		if strings.HasPrefix(path, "/swagger") {
			return c.Next()
		}

		start := time.Now()
		body := string(c.Body())
		if strings.Contains(path, "login") && body != "" {
			body = "[REDACTED]"
		}
		if len(body) > 2048 {
			body = body[:2048] + "...[truncated]"
		}

		err := c.Next()

		duration := time.Since(start).Milliseconds()
		status := c.Response().StatusCode()

		userID, _ := c.Locals("userId").(uint)
		username, _ := c.Locals("username").(string)
		if username == "" {
			username = "-"
		}

		log := adminmodel.OperationLog{
			UserID:    userID,
			Username:  username,
			Method:    c.Method(),
			Path:      path,
			Status:    status,
			Duration:  duration,
			IP:        c.IP(),
			UserAgent: c.Get("User-Agent"),
			Body:      body,
		}

		go store.DB.Create(&log)

		return err
	}
}
