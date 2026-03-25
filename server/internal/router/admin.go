package router

import (
	admin "base/internal/handler/admin"
	"base/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupAdmin registers all /admin/* routes (backend management).
func SetupAdmin(app *fiber.App) {
	g := app.Group("/admin")

	// Public auth routes
	auth := g.Group("/auth")
	auth.Post("/login", admin.Login)
	auth.Post("/logout", admin.Logout)
	auth.Post("/refresh", admin.RefreshToken)

	// Protected routes
	protected := g.Group("", middleware.JWTAuth(), middleware.CasbinAuth())

	// Operation log middleware (records POST/PUT/DELETE)
	protected.Use(middleware.OperationLog())

	// Auth info
	protected.Get("/auth/codes", admin.GetAccessCodes)
	protected.Post("/auth/change-password", admin.ChangePassword)

	// User info
	protected.Get("/user/info", admin.GetUserInfo)

	// Menu routes (for frontend routing)
	protected.Get("/menu/all", admin.GetAllMenus)

	// System management
	sys := protected.Group("/system")

	// Roles
	sys.Get("/role/all", admin.GetAllRoles)
	sys.Get("/role/list", admin.GetRoleList)
	sys.Post("/role", admin.CreateRole)
	sys.Put("/role/:id", admin.UpdateRole)
	sys.Delete("/role/:id", admin.DeleteRole)

	// Menus management
	sys.Get("/menu/list", admin.GetMenuList)
	sys.Get("/menu/name-exists", admin.CheckMenuNameExists)
	sys.Get("/menu/path-exists", admin.CheckMenuPathExists)
	sys.Post("/menu", admin.CreateMenu)
	sys.Put("/menu/:id", admin.UpdateMenu)
	sys.Delete("/menu/:id", admin.DeleteMenu)

	// Departments
	sys.Get("/dept/list", admin.GetDeptList)
	sys.Post("/dept", admin.CreateDept)
	sys.Put("/dept/:id", admin.UpdateDept)
	sys.Delete("/dept/:id", admin.DeleteDept)

	// Users management
	sys.Get("/user/list", admin.GetUserList)
	sys.Post("/user", admin.CreateUser)
	sys.Put("/user/:id", admin.UpdateUser)
	sys.Delete("/user/:id", admin.DeleteUser)

	// Configs management
	sys.Get("/config/list", admin.GetConfigList)
	sys.Get("/config/groups", admin.GetConfigGroups)
	sys.Post("/config", admin.CreateConfig)
	sys.Put("/config/:id", admin.UpdateConfig)
	sys.Delete("/config/:id", admin.DeleteConfig)

	// Operation logs
	sys.Get("/operation-log/list", admin.GetOperationLogList)
	sys.Delete("/operation-log/clear", admin.ClearOperationLog)
	sys.Delete("/operation-log/:id", admin.DeleteOperationLog)
}
