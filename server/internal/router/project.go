package router

import "github.com/gofiber/fiber/v2"

// SetupProject 注册下游项目的业务路由。
//
// 基底约定：基底仓库永不修改本文件（保持空实现），下游项目的路由全部注册在
// 这里，merge 基底更新时路由层不会产生冲突。需要 JWT/Casbin/操作日志时参考
// admin.go 中 protected 分组的中间件挂法。
func SetupProject(app *fiber.App) {
}
