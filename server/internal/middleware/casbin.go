package middleware

import (
	"log"

	adminmodel "base/internal/model/admin"
	"base/internal/store"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	casbinutil "github.com/casbin/casbin/v3/util"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gofiber/fiber/v2"
)

var Enforcer *casbin.Enforcer

type routePermission struct {
	Method string
	Path   string
	Code   string
}

var authenticatedRoutes = []routePermission{
	{Method: "GET", Path: "/admin/auth/codes"},
	{Method: "POST", Path: "/admin/auth/change-password"},
	{Method: "GET", Path: "/admin/user/info"},
	{Method: "GET", Path: "/admin/menu/all"},
}

var routePermissions = []routePermission{
	{Method: "GET", Path: "/admin/system/role/all", Code: "System:Role:List"},
	{Method: "GET", Path: "/admin/system/role/list", Code: "System:Role:List"},
	// 角色管理专用菜单树仅供"角色管理"页面使用，复用 System:Role:List 权限码：
	// 拥有角色管理列表权限即可调用，不强依赖"菜单管理"权限。
	{Method: "GET", Path: "/admin/system/role/menu-tree", Code: "System:Role:List"},
	{Method: "POST", Path: "/admin/system/role", Code: "System:Role:Create"},
	{Method: "PUT", Path: "/admin/system/role/:id", Code: "System:Role:Edit"},
	{Method: "DELETE", Path: "/admin/system/role/:id", Code: "System:Role:Delete"},
	{Method: "GET", Path: "/admin/system/menu/list", Code: "System:Menu:List"},
	{Method: "GET", Path: "/admin/system/menu/name-exists", Code: "System:Menu:List"},
	{Method: "GET", Path: "/admin/system/menu/path-exists", Code: "System:Menu:List"},
	{Method: "POST", Path: "/admin/system/menu", Code: "System:Menu:Create"},
	{Method: "PUT", Path: "/admin/system/menu/:id", Code: "System:Menu:Edit"},
	{Method: "DELETE", Path: "/admin/system/menu/:id", Code: "System:Menu:Delete"},
	{Method: "GET", Path: "/admin/system/dept/list", Code: "System:Dept:List"},
	{Method: "POST", Path: "/admin/system/dept", Code: "System:Dept:Create"},
	{Method: "PUT", Path: "/admin/system/dept/:id", Code: "System:Dept:Edit"},
	{Method: "DELETE", Path: "/admin/system/dept/:id", Code: "System:Dept:Delete"},
	{Method: "GET", Path: "/admin/system/user/list", Code: "System:User:List"},
	{Method: "POST", Path: "/admin/system/user", Code: "System:User:Create"},
	{Method: "PUT", Path: "/admin/system/user/:id", Code: "System:User:Edit"},
	{Method: "DELETE", Path: "/admin/system/user/:id", Code: "System:User:Delete"},
	{Method: "GET", Path: "/admin/system/config/list", Code: "System:Config:List"},
	{Method: "GET", Path: "/admin/system/config/groups", Code: "System:Config:List"},
	{Method: "POST", Path: "/admin/system/config", Code: "System:Config:Create"},
	{Method: "PUT", Path: "/admin/system/config/:id", Code: "System:Config:Edit"},
	{Method: "DELETE", Path: "/admin/system/config/:id", Code: "System:Config:Delete"},
	{Method: "GET", Path: "/admin/system/operation-log/list", Code: "System:OperationLog:List"},
	{Method: "DELETE", Path: "/admin/system/operation-log/clear", Code: "System:OperationLog:Delete"},
	{Method: "DELETE", Path: "/admin/system/operation-log/:id", Code: "System:OperationLog:Delete"},
}

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
		if isSuperAdminUser(c) {
			return c.Next()
		}

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

		if hasRole(roles, "super") {
			return c.Next()
		}

		if matchesRoute(authenticatedRoutes, act, obj) {
			return c.Next()
		}

		code, ok := permissionCodeForRoute(act, obj)
		if ok && roleHasPermissionCode(roles, code) {
			return c.Next()
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    -1,
			"data":    nil,
			"error":   "Forbidden",
			"message": "没有权限执行此操作",
		})
	}
}

func isSuperAdminUser(c *fiber.Ctx) bool {
	switch userID := c.Locals("userId").(type) {
	case uint:
		return userID == 1
	case uint64:
		return userID == 1
	case int:
		return userID == 1
	case int64:
		return userID == 1
	case float64:
		return userID == 1
	default:
		return false
	}
}

func hasRole(roles []string, code string) bool {
	for _, role := range roles {
		if role == code {
			return true
		}
	}
	return false
}

func matchesRoute(routes []routePermission, method string, path string) bool {
	for _, route := range routes {
		if route.Method == method && casbinutil.KeyMatch2(path, route.Path) {
			return true
		}
	}
	return false
}

func permissionCodeForRoute(method string, path string) (string, bool) {
	for _, route := range routePermissions {
		if route.Method == method && casbinutil.KeyMatch2(path, route.Path) {
			return route.Code, true
		}
	}
	return "", false
}

func roleHasPermissionCode(roles []string, code string) bool {
	if len(roles) == 0 || code == "" {
		return false
	}

	var count int64
	store.DB.Model(&adminmodel.Menu{}).
		Joins("JOIN role_menus ON role_menus.menu_id = sys_menus.id").
		Joins("JOIN sys_roles ON sys_roles.id = role_menus.role_id").
		Where("sys_roles.code IN ? AND sys_roles.status = ? AND sys_roles.deleted_at IS NULL", roles, 1).
		Where("sys_menus.auth_code = ? AND sys_menus.status = ?", code, 1).
		Count(&count)
	return count > 0
}
