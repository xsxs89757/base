// Package store 是数据层在本项目里的挂载点。
//
// 真正的数据层实现在 github.com/xsxs89757/base-kit/store：连接、迁移、基底种子数据都在那边。
// 本包只提供两样东西：
//   - 把 project.go 里登记的模型和种子数据交给 kit（ProjectModels / ProjectSeed）
//   - 给 project.go 的函数体提供 DB、syncSeedMenu、refreshRoleMenus，让挂载点文件保持原样
//
// 新写的业务代码请直接 import kit 的 store（`kitstore "github.com/xsxs89757/base-kit/store"`），
// 用 kitstore.DB、kitstore.IsUniqueViolation、kitstore.SyncSeedMenus 等；
// 本包的这些同名符号只为兼容 project.go 里已有的写法而存在。
package store

import (
	adminmodel "github.com/xsxs89757/base-kit/model/admin"
	kitstore "github.com/xsxs89757/base-kit/store"

	"gorm.io/gorm"
)

// DB 指向 kit 打开的同一个连接，在 ProjectSeed 执行前由 kit 赋值。
// 只在 projectSeed 及其之后的运行期可用（HTTP 请求进来时早已就绪）。
var DB *gorm.DB

// ProjectModels 返回 project.go 里登记的下游模型，交给 kit 一起 AutoMigrate。
func ProjectModels() []any {
	return projectModels
}

// ProjectSeed 由 kit 在基底种子数据之后调用。
func ProjectSeed(db *gorm.DB) {
	DB = db
	projectSeed()
}

// syncSeedMenu / refreshRoleMenus 是 project.go 文档里提到的助手，转调 kit 的实现。
// 保留这两个未导出的名字，是为了让下游已经写好的 projectSeed 函数体一个字都不用改。
func syncSeedMenu(menu adminmodel.Menu, parentName string) (adminmodel.Menu, bool) {
	return kitstore.SyncSeedMenu(menu, parentName)
}

func refreshRoleMenus() {
	kitstore.RefreshRoleMenus()
}
