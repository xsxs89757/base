package middleware

import (
	"log"

	"base/internal/store"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gofiber/fiber/v2"
)

var Enforcer *casbin.Enforcer

func InitCasbin() {
	adapter, err := gormadapter.NewAdapterByDB(store.DB)
	if err != nil {
		log.Fatalf("failed to create casbin adapter: %v", err)
	}

	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act || r.sub == "super"
`)
	if err != nil {
		log.Fatalf("failed to create casbin model: %v", err)
	}

	Enforcer, err = casbin.NewEnforcer(m, adapter)
	if err != nil {
		log.Fatalf("failed to create casbin enforcer: %v", err)
	}

	if err = Enforcer.LoadPolicy(); err != nil {
		log.Fatalf("failed to load casbin policy: %v", err)
	}

	seedPolicies()
}

func seedPolicies() {
	policies, _ := Enforcer.GetPolicy()
	if len(policies) > 0 {
		return
	}

	log.Println("seeding casbin policies...")

	// admin role policies
	_, _ = Enforcer.AddPolicies([][]string{
		{"admin", "/admin/user/*", "GET"},
		{"admin", "/admin/menu/*", "GET"},
		{"admin", "/admin/auth/*", "GET"},
		{"admin", "/admin/auth/*", "POST"},
		{"admin", "/admin/system/*", "GET"},
		{"admin", "/admin/system/*", "POST"},
		{"admin", "/admin/system/*", "PUT"},
		{"admin", "/admin/system/*", "DELETE"},
	})

	// user role policies (read only)
	_, _ = Enforcer.AddPolicies([][]string{
		{"user", "/admin/user/*", "GET"},
		{"user", "/admin/menu/*", "GET"},
		{"user", "/admin/auth/*", "GET"},
		{"user", "/admin/auth/*", "POST"},
	})

	_ = Enforcer.SavePolicy()
	log.Println("casbin policies seeded")
}

func CasbinAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roles, ok := c.Locals("roles").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    -1,
				"data":    nil,
				"error":   "Forbidden",
				"message": "Forbidden",
			})
		}

		obj := c.Path()
		act := c.Method()

		for _, role := range roles {
			if role == "super" {
				return c.Next()
			}
			allowed, err := Enforcer.Enforce(role, obj, act)
			if err != nil {
				continue
			}
			if allowed {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    -1,
			"data":    nil,
			"error":   "Forbidden",
			"message": "没有权限执行此操作",
		})
	}
}
