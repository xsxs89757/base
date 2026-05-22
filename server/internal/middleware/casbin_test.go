package middleware

import (
	"net/http"
	"testing"

	adminmodel "base/internal/model/admin"
	"base/internal/store"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCasbinAuthTestDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&adminmodel.Role{}, &adminmodel.Menu{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	store.DB = db
	InitCasbin()
}

func grantMenuCode(t *testing.T, roleCode string, authCode string) {
	t.Helper()

	role := adminmodel.Role{Name: roleCode, Code: roleCode, Status: 1}
	if err := store.DB.Create(&role).Error; err != nil {
		t.Fatalf("create role %s: %v", roleCode, err)
	}
	menu := adminmodel.Menu{Name: authCode, Type: "menu", Title: authCode, AuthCode: authCode, Status: 1}
	if err := store.DB.Create(&menu).Error; err != nil {
		t.Fatalf("create menu %s: %v", authCode, err)
	}
	if err := store.DB.Model(&role).Association("Menus").Replace([]adminmodel.Menu{menu}); err != nil {
		t.Fatalf("grant menu code %s: %v", authCode, err)
	}
}

func testAppWithRoles(roles []string) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("roles", roles)
		return c.Next()
	})
	app.Use(CasbinAuth())
	app.Get("/admin/system/user/list", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Delete("/admin/system/user/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

func TestCasbinAuthAllowsUserIDOneAsSuperAdmin(t *testing.T) {
	setupCasbinAuthTestDB(t)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", uint(1))
		return c.Next()
	})
	app.Use(CasbinAuth())
	app.Delete("/admin/system/user/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req, err := http.NewRequest(http.MethodDelete, "/admin/system/user/2", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestCasbinAuthAllowsCustomRoleByGrantedMenuAuthCode(t *testing.T) {
	setupCasbinAuthTestDB(t)
	grantMenuCode(t, "auditor", "System:User:List")

	req, err := http.NewRequest(http.MethodGet, "/admin/system/user/list", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := testAppWithRoles([]string{"auditor"}).Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestCasbinAuthDeniesAdminRouteWithoutMatchingMenuAuthCode(t *testing.T) {
	setupCasbinAuthTestDB(t)
	grantMenuCode(t, "admin", "System:User:List")

	req, err := http.NewRequest(http.MethodDelete, "/admin/system/user/2", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := testAppWithRoles([]string{"admin"}).Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}
