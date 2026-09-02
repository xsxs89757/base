package router

import "github.com/gofiber/fiber/v2"

// Setup 注册本项目自己的路由，由 main.go 通过 basekit.Options.Routes 传给 kit，
// 在 kit 的 /admin 管理接口之后执行。
//
// 需要覆盖 kit 某个接口时用 basekit.Options.PreRoutes（Fiber 先注册先匹配）。
func Setup(app *fiber.App) {
	SetupAPI(app)
	SetupProject(app)
}
