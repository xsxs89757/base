package admin

import (
	"strconv"

	"base/internal/dto"
	admindto "base/internal/dto/admin"
	adminsvc "base/internal/service/admin"
	"base/internal/validator"

	"github.com/gofiber/fiber/v2"
)

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 分页查询用户列表，支持按用户名和状态筛选
// @Tags 系统管理 - 用户
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param username query string false "用户名(模糊搜索)"
// @Param status query int false "状态: 0=禁用 1=启用"
// @Success 200 {object} dto.Response{data=dto.PageData{items=[]admindto.UserItem}}
// @Failure 401 {object} dto.Response
// @Router /admin/system/user/list [get]
func GetUserList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	params := adminsvc.UserListParams{
		Page:     page,
		PageSize: pageSize,
		Username: c.Query("username"),
	}

	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		params.Status = &v
	}

	users, total, err := adminsvc.GetUserList(params)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to get users")
	}

	items := make([]admindto.UserItem, len(users))
	for i, u := range users {
		roles := make([]string, len(u.Roles))
		roleIDs := make([]uint, len(u.Roles))
		for j, r := range u.Roles {
			roles[j] = r.Name
			roleIDs[j] = r.ID
		}
		items[i] = admindto.UserItem{
			ID:       u.ID,
			Username: u.Username,
			RealName: u.RealName,
			Email:    u.Email,
			Phone:    u.Phone,
			Status:   u.Status,
			Roles:    roles,
			RoleIDs:  roleIDs,
			Remark:   u.Remark,
		}
	}
	return dto.PageSuccess(c, items, total)
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户并分配角色
// @Tags 系统管理 - 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body admindto.CreateUserRequest true "用户信息"
// @Success 200 {object} dto.Response{data=dto.IDResponse}
// @Failure 400 {object} dto.Response
// @Router /admin/system/user [post]
func CreateUser(c *fiber.Ctx) error {
	var req admindto.CreateUserRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	user := adminsvc.NewUser(req.Username, req.Password, req.RealName, req.Email, req.Phone, req.Status, req.Remark)
	if err := adminsvc.CreateUser(user, req.RoleIDs); err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to create user: "+err.Error())
	}
	return dto.Success(c, fiber.Map{"id": user.ID})
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息和角色分配
// @Tags 系统管理 - 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body admindto.UpdateUserRequest true "用户信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /admin/system/user/{id} [put]
func UpdateUser(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var req admindto.UpdateUserRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	updates := map[string]any{
		"real_name": req.RealName,
		"email":     req.Email,
		"phone":     req.Phone,
		"status":    req.Status,
		"remark":    req.Remark,
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}

	if err := adminsvc.UpdateUser(uint(id), updates, req.RoleIDs); err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to update user")
	}
	return dto.Success(c, nil)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户
// @Tags 系统管理 - 用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} dto.Response
// @Router /admin/system/user/{id} [delete]
func DeleteUser(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := adminsvc.DeleteUser(uint(id)); err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to delete user")
	}
	return dto.Success(c, nil)
}
