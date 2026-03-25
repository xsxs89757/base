package admin

import "base/internal/model"

type User struct {
	model.BaseModel
	Username string `json:"username" gorm:"uniqueIndex;size:64;not null"`
	Password string `json:"-" gorm:"size:128;not null"`
	RealName string `json:"realName" gorm:"size:64"`
	Avatar   string `json:"avatar" gorm:"size:256"`
	Email    string `json:"email" gorm:"size:128"`
	Phone    string `json:"phone" gorm:"size:20"`
	Status   int    `json:"status" gorm:"default:1;comment:0=disabled 1=enabled"`
	HomePath string `json:"homePath,omitempty" gorm:"size:128"`
	Remark   string `json:"remark" gorm:"size:256"`
	Roles    []Role `json:"roles" gorm:"many2many:user_roles;"`
}

func (User) TableName() string {
	return "sys_users"
}

type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
