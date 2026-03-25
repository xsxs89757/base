package store

import (
	"log"

	"base/config"
	adminmodel "base/internal/model/admin"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	var err error
	cfg := config.C.Database

	var logLevel logger.LogLevel
	if config.C.Server.Mode == "development" {
		logLevel = logger.Info
	} else {
		logLevel = logger.Warn
	}

	DB, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// AutoMigrate: 无表自动建表，新字段自动加列（不删列不改类型，和 FreeSql 行为一致）
	if err = DB.AutoMigrate(
		&adminmodel.User{},
		&adminmodel.Role{},
		&adminmodel.Menu{},
		&adminmodel.Dept{},
		&adminmodel.Config{},
		&adminmodel.OperationLog{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	seed()
}

// ---------------------------------------------------------------------------
// 增量种子：每条数据按唯一键判断，已存在则跳过，不存在则插入
// 新增模块只需在这里加定义，重启即可自动补齐，无需删库
// ---------------------------------------------------------------------------

func seed() {
	seedRoles()
	seedMenus()
	seedUsers()
	seedDepts()
	seedConfigs()
	log.Println("seed check completed")
}

// --- Roles ---

func seedRoles() {
	roles := []adminmodel.Role{
		{Name: "超级管理员", Code: "super", Status: 1, Remark: "拥有所有权限"},
		{Name: "管理员", Code: "admin", Status: 1, Remark: "普通管理权限"},
		{Name: "普通用户", Code: "user", Status: 1, Remark: "基础查看权限"},
	}
	for _, r := range roles {
		var exists adminmodel.Role
		if DB.Where("code = ?", r.Code).First(&exists).Error != nil {
			DB.Create(&r)
			log.Printf("  [seed] role created: %s", r.Code)
		}
	}
}

// --- Menus ---

type menuDef struct {
	adminmodel.Menu
	ParentName string
	Buttons    []adminmodel.Menu
}

func seedMenus() {
	defs := []menuDef{
		{Menu: adminmodel.Menu{Name: "Dashboard", Path: "/dashboard", Type: "catalog", Icon: "lucide:layout-dashboard", Title: "page.dashboard.title", OrderNo: -1, Status: 1}},
		{Menu: adminmodel.Menu{Name: "Analytics", Path: "/analytics", Component: "/dashboard/analytics/index", Type: "menu", Icon: "lucide:area-chart", Title: "page.dashboard.analytics", OrderNo: 1, Status: 1, AffixTab: true}, ParentName: "Dashboard"},
		{Menu: adminmodel.Menu{Name: "Workspace", Path: "/workspace", Component: "/dashboard/workspace/index", Type: "menu", Icon: "carbon:workspace", Title: "page.dashboard.workspace", OrderNo: 2, Status: 1}, ParentName: "Dashboard"},
		{Menu: adminmodel.Menu{Name: "System", Path: "/system", Type: "catalog", Icon: "carbon:settings", Title: "system.title", OrderNo: 9997, Status: 1}},
		{
			Menu:       adminmodel.Menu{Name: "SystemUser", Path: "/system/user", Component: "/system/user/list", Type: "menu", Icon: "mdi:account-outline", Title: "system.user.title", OrderNo: 1, Status: 1, AuthCode: "System:User:List"},
			ParentName: "System",
			Buttons: []adminmodel.Menu{
				{Name: "SystemUserCreate", Type: "button", Title: "common.create", AuthCode: "System:User:Create", Status: 1},
				{Name: "SystemUserEdit", Type: "button", Title: "common.edit", AuthCode: "System:User:Edit", Status: 1},
				{Name: "SystemUserDelete", Type: "button", Title: "common.delete", AuthCode: "System:User:Delete", Status: 1},
			},
		},
		{
			Menu:       adminmodel.Menu{Name: "SystemRole", Path: "/system/role", Component: "/system/role/list", Type: "menu", Icon: "mdi:account-group", Title: "system.role.title", OrderNo: 2, Status: 1, AuthCode: "System:Role:List"},
			ParentName: "System",
			Buttons: []adminmodel.Menu{
				{Name: "SystemRoleCreate", Type: "button", Title: "common.create", AuthCode: "System:Role:Create", Status: 1},
				{Name: "SystemRoleEdit", Type: "button", Title: "common.edit", AuthCode: "System:Role:Edit", Status: 1},
				{Name: "SystemRoleDelete", Type: "button", Title: "common.delete", AuthCode: "System:Role:Delete", Status: 1},
			},
		},
		{
			Menu:       adminmodel.Menu{Name: "SystemMenu", Path: "/system/menu", Component: "/system/menu/list", Type: "menu", Icon: "carbon:menu", Title: "system.menu.title", OrderNo: 3, Status: 1, AuthCode: "System:Menu:List"},
			ParentName: "System",
			Buttons: []adminmodel.Menu{
				{Name: "SystemMenuCreate", Type: "button", Title: "common.create", AuthCode: "System:Menu:Create", Status: 1},
				{Name: "SystemMenuEdit", Type: "button", Title: "common.edit", AuthCode: "System:Menu:Edit", Status: 1},
				{Name: "SystemMenuDelete", Type: "button", Title: "common.delete", AuthCode: "System:Menu:Delete", Status: 1},
			},
		},
		{
			Menu:       adminmodel.Menu{Name: "SystemDept", Path: "/system/dept", Component: "/system/dept/list", Type: "menu", Icon: "carbon:container-services", Title: "system.dept.title", OrderNo: 4, Status: 1, AuthCode: "System:Dept:List"},
			ParentName: "System",
			Buttons: []adminmodel.Menu{
				{Name: "SystemDeptCreate", Type: "button", Title: "common.create", AuthCode: "System:Dept:Create", Status: 1},
				{Name: "SystemDeptEdit", Type: "button", Title: "common.edit", AuthCode: "System:Dept:Edit", Status: 1},
				{Name: "SystemDeptDelete", Type: "button", Title: "common.delete", AuthCode: "System:Dept:Delete", Status: 1},
			},
		},
		{
			Menu:       adminmodel.Menu{Name: "SystemConfig", Path: "/system/config", Component: "/system/config/list", Type: "menu", Icon: "carbon:settings-adjust", Title: "system.config.title", OrderNo: 5, Status: 1, AuthCode: "System:Config:List"},
			ParentName: "System",
			Buttons: []adminmodel.Menu{
				{Name: "SystemConfigCreate", Type: "button", Title: "common.create", AuthCode: "System:Config:Create", Status: 1},
				{Name: "SystemConfigEdit", Type: "button", Title: "common.edit", AuthCode: "System:Config:Edit", Status: 1},
				{Name: "SystemConfigDelete", Type: "button", Title: "common.delete", AuthCode: "System:Config:Delete", Status: 1},
			},
		},
		{
			Menu:       adminmodel.Menu{Name: "SystemOperationLog", Path: "/system/operation-log", Component: "/system/operation-log/list", Type: "menu", Icon: "carbon:activity", Title: "system.operationLog.title", OrderNo: 6, Status: 1, AuthCode: "System:OperationLog:List"},
			ParentName: "System",
			Buttons: []adminmodel.Menu{
				{Name: "SystemOperationLogDelete", Type: "button", Title: "common.delete", AuthCode: "System:OperationLog:Delete", Status: 1},
			},
		},
		{Menu: adminmodel.Menu{Name: "About", Path: "/about", Component: "_core/about/index", Type: "menu", Icon: "lucide:copyright", Title: "demos.vben.about", OrderNo: 9999, Status: 1}},
	}

	newMenuCreated := false
	for _, d := range defs {
		var exists adminmodel.Menu
		if DB.Where("name = ?", d.Name).First(&exists).Error == nil {
			continue
		}
		if d.ParentName != "" {
			var parent adminmodel.Menu
			if DB.Where("name = ?", d.ParentName).First(&parent).Error == nil {
				d.Menu.ParentID = parent.ID
			}
		}
		DB.Create(&d.Menu)
		log.Printf("  [seed] menu created: %s", d.Name)
		newMenuCreated = true

		for _, btn := range d.Buttons {
			var btnExists adminmodel.Menu
			if DB.Where("name = ?", btn.Name).First(&btnExists).Error == nil {
				continue
			}
			btn.ParentID = d.Menu.ID
			DB.Create(&btn)
			log.Printf("  [seed]   button created: %s", btn.Name)
		}
	}

	if newMenuCreated {
		refreshRoleMenus()
	}
}

func refreshRoleMenus() {
	var superRole, adminRole adminmodel.Role
	if DB.Where("code = ?", "super").First(&superRole).Error != nil {
		return
	}

	var allMenus []adminmodel.Menu
	DB.Find(&allMenus)
	DB.Model(&superRole).Association("Menus").Replace(allMenus)
	log.Println("  [seed] super role menus refreshed")

	if DB.Where("code = ?", "admin").First(&adminRole).Error == nil {
		var adminMenus []adminmodel.Menu
		DB.Where("type != ?", "button").Find(&adminMenus)
		DB.Model(&adminRole).Association("Menus").Replace(adminMenus)
		log.Println("  [seed] admin role menus refreshed")
	}
}

// --- Users ---

func seedUsers() {
	userDefs := []struct {
		Username string
		RealName string
		RoleCode string
		HomePath string
	}{
		{"vben", "Vben", "super", ""},
		{"admin", "Admin", "admin", "/workspace"},
		{"jack", "Jack", "user", "/analytics"},
	}

	for _, u := range userDefs {
		var exists adminmodel.User
		if DB.Where("username = ?", u.Username).First(&exists).Error == nil {
			continue
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		var role adminmodel.Role
		DB.Where("code = ?", u.RoleCode).First(&role)
		user := adminmodel.User{
			Username: u.Username,
			Password: string(hash),
			RealName: u.RealName,
			Status:   1,
			HomePath: u.HomePath,
			Roles:    []adminmodel.Role{role},
		}
		DB.Create(&user)
		log.Printf("  [seed] user created: %s", u.Username)
	}
}

// --- Departments ---

func seedDepts() {
	type deptDef struct {
		Name       string
		OrderNo    int
		ParentName string
	}
	defs := []deptDef{
		{"总公司", 1, ""},
		{"技术部", 1, "总公司"},
		{"市场部", 2, "总公司"},
		{"财务部", 3, "总公司"},
	}
	for _, d := range defs {
		var exists adminmodel.Dept
		if DB.Where("name = ?", d.Name).First(&exists).Error == nil {
			continue
		}
		dept := adminmodel.Dept{Name: d.Name, OrderNo: d.OrderNo, Status: 1}
		if d.ParentName != "" {
			var parent adminmodel.Dept
			if DB.Where("name = ?", d.ParentName).First(&parent).Error == nil {
				dept.ParentID = parent.ID
			}
		}
		DB.Create(&dept)
		log.Printf("  [seed] dept created: %s", d.Name)
	}
}

// --- Configs ---

func seedConfigs() {
	defs := []adminmodel.Config{
		{ConfigKey: "site_name", ConfigValue: "Admin 后台管理系统", ConfigGroup: "basic", Status: 1, Remark: "站点名称"},
		{ConfigKey: "site_logo", ConfigValue: "/logo.png", ConfigGroup: "basic", Status: 1, Remark: "站点 Logo"},
		{ConfigKey: "upload_max_size", ConfigValue: "10", ConfigGroup: "upload", Status: 1, Remark: "上传文件最大大小(MB)"},
		{ConfigKey: "login_captcha", ConfigValue: "false", ConfigGroup: "security", Status: 1, Remark: "登录是否需要验证码"},
		{ConfigKey: "password_min_length", ConfigValue: "6", ConfigGroup: "security", Status: 1, Remark: "密码最小长度"},
	}
	for _, cfg := range defs {
		var exists adminmodel.Config
		if DB.Where("config_key = ?", cfg.ConfigKey).First(&exists).Error == nil {
			continue
		}
		DB.Create(&cfg)
		log.Printf("  [seed] config created: %s", cfg.ConfigKey)
	}
}
