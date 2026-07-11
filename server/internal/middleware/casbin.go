package middleware

import (
	"log"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	adminmodel "base/internal/model/admin"
	"base/internal/store"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gofiber/fiber/v2"
)

var Enforcer *casbin.Enforcer

type routePermission struct {
	Method string
	Path   string
	Code   string
	re     *regexp.Regexp // 由 init 预编译；casbin 的 KeyMatch2 每次调用都现编正则，热路径上白耗 CPU
}

// compileRoutePattern 把 "/admin/system/user/:id" 编译成 ^/admin/system/user/[^/]+$，
// 与 KeyMatch2 对 :param 段的匹配语义一致（路由表里只用了字面段和 :param 两种形式）。
func compileRoutePattern(path string) *regexp.Regexp {
	segs := strings.Split(path, "/")
	for i, seg := range segs {
		if strings.HasPrefix(seg, ":") && len(seg) > 1 {
			segs[i] = "[^/]+"
		} else {
			segs[i] = regexp.QuoteMeta(seg)
		}
	}
	return regexp.MustCompile("^" + strings.Join(segs, "/") + "$")
}

func init() {
	for i := range authenticatedRoutes {
		authenticatedRoutes[i].re = compileRoutePattern(authenticatedRoutes[i].Path)
	}
	for i := range routePermissions {
		routePermissions[i].re = compileRoutePattern(routePermissions[i].Path)
	}
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

	// 换库/重新初始化后旧缓存必须作废（测试里会反复 InitCasbin）
	InvalidatePermissionCache()

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

// OperatorIsSuper 判断当前请求者是否具备超级管理员身份：内置超管 user（id=1）或
// 持有 code=super 的角色。供角色/用户管理 handler 用来限制"授予 super 角色"这类高危操作，
// 防止普通管理员把自己或他人提升到不受权限体系约束的 super 层。
func OperatorIsSuper(c *fiber.Ctx) bool {
	if isSuperAdminUser(c) {
		return true
	}
	roles, _ := c.Locals("roles").([]string)
	return hasRole(roles, "super")
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
		if route.Method == method && route.re.MatchString(path) {
			return true
		}
	}
	return false
}

func permissionCodeForRoute(method string, path string) (string, bool) {
	for _, route := range routePermissions {
		if route.Method == method && route.re.MatchString(path) {
			return route.Code, true
		}
	}
	return "", false
}

// 权限码查询缓存：每个受权限保护的请求都要做一次 3 表 JOIN，
// 实测把接口吞吐拉低近一半（9.7k -> 5.6k req/s）。角色-菜单映射是低频变更数据，
// 用带 TTL 的进程内缓存，角色/菜单变更时由 handler 调 InvalidatePermissionCache 立即失效。
const permCacheTTL = time.Minute

type permCacheEntry struct {
	allowed   bool
	expiresAt time.Time
}

var (
	permCacheMu sync.RWMutex
	permCache   = map[string]permCacheEntry{}
)

func permCacheKey(roles []string, code string) string {
	sorted := append([]string(nil), roles...)
	sort.Strings(sorted)
	return strings.Join(sorted, ",") + "|" + code
}

// InvalidatePermissionCache 清空权限码缓存。
// 角色、菜单的增删改都会改变角色-菜单映射，对应 handler 必须调用本函数，
// 保证权限调整即时生效，而不是等 TTL 过期。
func InvalidatePermissionCache() {
	permCacheMu.Lock()
	permCache = map[string]permCacheEntry{}
	permCacheMu.Unlock()
}

func roleHasPermissionCode(roles []string, code string) bool {
	if len(roles) == 0 || code == "" {
		return false
	}

	key := permCacheKey(roles, code)
	permCacheMu.RLock()
	entry, ok := permCache[key]
	permCacheMu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return entry.allowed
	}

	var count int64
	err := store.DB.Model(&adminmodel.Menu{}).
		Joins("JOIN role_menus ON role_menus.menu_id = sys_menus.id").
		Joins("JOIN sys_roles ON sys_roles.id = role_menus.role_id").
		Where("sys_roles.code IN ? AND sys_roles.status = ? AND sys_roles.deleted_at IS NULL", roles, 1).
		Where("sys_menus.auth_code = ? AND sys_menus.status = ?", code, 1).
		Count(&count).Error
	if err != nil {
		// 查询失败按无权限处理，但不缓存，避免一次抖动把用户锁一分钟
		log.Printf("[casbin] permission query failed: %v", err)
		return false
	}

	allowed := count > 0
	permCacheMu.Lock()
	permCache[key] = permCacheEntry{allowed: allowed, expiresAt: time.Now().Add(permCacheTTL)}
	permCacheMu.Unlock()
	return allowed
}
