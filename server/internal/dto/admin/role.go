package admin

// CreateRoleRequest 创建/更新角色请求
type CreateRoleRequest struct {
	Name    string `json:"name" validate:"required" example:"新角色"`
	Code    string `json:"code" validate:"required" example:"new_role"`
	Status  int    `json:"status" validate:"oneof=0 1" example:"1"`
	Remark  string `json:"remark"`
	MenuIDs []uint `json:"menuIds"`
}

// RoleItem 角色列表项
type RoleItem struct {
	ID          uint   `json:"id" example:"1"`
	Name        string `json:"name" example:"超级管理员"`
	Code        string `json:"code" example:"super"`
	Status      int    `json:"status" example:"1"`
	Remark      string `json:"remark" example:"拥有所有权限"`
	Permissions []uint `json:"permissions"`
	CreateTime  string `json:"createTime" example:"2026/01/01 00:00:00"`
}
