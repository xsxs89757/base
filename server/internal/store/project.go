package store

// 基底约定：基底仓库永不修改本文件（保持空实现），下游项目的模型登记与
// 种子数据全部放在这里，merge 基底更新时不会与 store.go 冲突。

// projectModels 中的模型会与基底模型一起 AutoMigrate，例如：
//
//	var projectModels = []any{&bizmodel.Order{}, &bizmodel.Product{}}
var projectModels []any

// projectSeed 在基底 seed() 之后执行。菜单用 syncSeedMenu 追加，追加后调用
// refreshRoleMenus()；其他数据参考 store.go 中 seedXxx 的「按唯一键判断、
// 存在即跳过」写法。
func projectSeed() {}
