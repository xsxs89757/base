package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	adminmodel "base/internal/model/admin"
	"base/internal/store"
	"base/internal/validator"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRolePermissionTestDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&adminmodel.User{}, &adminmodel.Role{}, &adminmodel.Menu{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	store.DB = db
	validator.Init()
}

func TestUpdateRoleAcceptsPermissionsPayload(t *testing.T) {
	setupRolePermissionTestDB(t)

	initialMenu := adminmodel.Menu{Name: "InitialMenu", Type: "menu", Title: "initial", Status: 1}
	targetMenu := adminmodel.Menu{Name: "TargetMenu", Type: "menu", Title: "target", Status: 1}
	if err := store.DB.Create(&initialMenu).Error; err != nil {
		t.Fatalf("create initial menu: %v", err)
	}
	if err := store.DB.Create(&targetMenu).Error; err != nil {
		t.Fatalf("create target menu: %v", err)
	}

	role := adminmodel.Role{Name: "Auditor", Code: "auditor", Status: 1}
	if err := store.DB.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := store.DB.Model(&role).Association("Menus").Replace([]adminmodel.Menu{initialMenu}); err != nil {
		t.Fatalf("seed role menus: %v", err)
	}

	body, err := json.Marshal(fiber.Map{
		"name":        "Auditor",
		"code":        "auditor",
		"status":      1,
		"permissions": []uint{targetMenu.ID},
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	app := fiber.New()
	app.Put("/role/:id", UpdateRole)
	req, err := http.NewRequest(http.MethodPut, "/role/1", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("update role request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var updated adminmodel.Role
	if err := store.DB.Preload("Menus").First(&updated, role.ID).Error; err != nil {
		t.Fatalf("load updated role: %v", err)
	}
	if len(updated.Menus) != 1 || updated.Menus[0].ID != targetMenu.ID {
		t.Fatalf("expected role menus [%d], got %#v", targetMenu.ID, updated.Menus)
	}
}

func TestGetAllMenusIncludesAncestorsForGrantedChildMenu(t *testing.T) {
	setupRolePermissionTestDB(t)

	parent := adminmodel.Menu{Name: "System", Path: "/system", Type: "catalog", Title: "system.title", Status: 1}
	if err := store.DB.Create(&parent).Error; err != nil {
		t.Fatalf("create parent menu: %v", err)
	}
	child := adminmodel.Menu{Name: "SystemUser", ParentID: parent.ID, Path: "/system/user", Component: "/system/user/list", Type: "menu", Title: "system.user.title", Status: 1}
	if err := store.DB.Create(&child).Error; err != nil {
		t.Fatalf("create child menu: %v", err)
	}

	role := adminmodel.Role{Name: "Limited", Code: "limited", Status: 1}
	if err := store.DB.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := store.DB.Model(&role).Association("Menus").Replace([]adminmodel.Menu{child}); err != nil {
		t.Fatalf("seed role menus: %v", err)
	}

	super := adminmodel.User{Username: "super", Password: "unused", Status: 1}
	if err := store.DB.Create(&super).Error; err != nil {
		t.Fatalf("create super placeholder: %v", err)
	}
	user := adminmodel.User{Username: "limited", Password: "unused", Status: 1, Roles: []adminmodel.Role{role}}
	if err := store.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	app := fiber.New()
	app.Get("/menu/all", func(c *fiber.Ctx) error {
		c.Locals("username", "limited")
		return GetAllMenus(c)
	})
	req, err := http.NewRequest(http.MethodGet, "/menu/all", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("get menus request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var parsed struct {
		Code int `json:"code"`
		Data []struct {
			Name     string `json:"name"`
			Children []struct {
				Name string `json:"name"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(parsed.Data) != 1 || parsed.Data[0].Name != "System" {
		t.Fatalf("expected System parent in menu tree, got %#v", parsed.Data)
	}
	if len(parsed.Data[0].Children) != 1 || parsed.Data[0].Children[0].Name != "SystemUser" {
		t.Fatalf("expected SystemUser child in menu tree, got %#v", parsed.Data[0].Children)
	}
}

func TestGetUserInfoFallsBackToFirstAccessibleMenuWhenHomePathIsNotGranted(t *testing.T) {
	setupRolePermissionTestDB(t)

	parent := adminmodel.Menu{Name: "System", Path: "/system", Type: "catalog", Title: "system.title", Status: 1}
	if err := store.DB.Create(&parent).Error; err != nil {
		t.Fatalf("create parent menu: %v", err)
	}
	child := adminmodel.Menu{Name: "SystemUser", ParentID: parent.ID, Path: "/system/user", Component: "/system/user/list", Type: "menu", Title: "system.user.title", Status: 1}
	if err := store.DB.Create(&child).Error; err != nil {
		t.Fatalf("create child menu: %v", err)
	}

	role := adminmodel.Role{Name: "Limited", Code: "limited", Status: 1}
	if err := store.DB.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := store.DB.Model(&role).Association("Menus").Replace([]adminmodel.Menu{child}); err != nil {
		t.Fatalf("seed role menus: %v", err)
	}

	super := adminmodel.User{Username: "super", Password: "unused", Status: 1}
	if err := store.DB.Create(&super).Error; err != nil {
		t.Fatalf("create super placeholder: %v", err)
	}
	user := adminmodel.User{Username: "limited", Password: "unused", Status: 1, HomePath: "/analytics", Roles: []adminmodel.Role{role}}
	if err := store.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	app := fiber.New()
	app.Get("/user/info", func(c *fiber.Ctx) error {
		c.Locals("userId", user.ID)
		return GetUserInfo(c)
	})
	req, err := http.NewRequest(http.MethodGet, "/user/info", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("get user info request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var parsed struct {
		Data struct {
			HomePath string `json:"homePath"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if parsed.Data.HomePath != "/system/user" {
		t.Fatalf("expected fallback home path /system/user, got %q", parsed.Data.HomePath)
	}
}
