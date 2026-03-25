package admin

import (
	"strconv"

	"base/internal/dto"
	admindto "base/internal/dto/admin"
	adminmodel "base/internal/model/admin"
	"base/internal/store"
	"base/internal/validator"

	"github.com/gofiber/fiber/v2"
)

// GetAllRoles 获取全部角色（简单列表，不分页）
// @Summary 获取全部角色
// @Tags 系统管理 - 角色
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]admindto.RoleItem}
// @Router /admin/system/role/all [get]
func GetAllRoles(c *fiber.Ctx) error {
	var roles []adminmodel.Role
	store.DB.Where("status = ?", 1).Order("id").Find(&roles)

	items := make([]admindto.RoleItem, len(roles))
	for i, r := range roles {
		items[i] = admindto.RoleItem{
			ID:   r.ID,
			Name: r.Name,
			Code: r.Code,
		}
	}
	return dto.Success(c, items)
}

// GetRoleList 获取角色列表
// @Summary 获取角色列表
// @Description 分页查询角色列表，支持按名称和状态筛选
// @Tags 系统管理 - 角色
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param name query string false "角色名(模糊搜索)"
// @Param status query string false "状态: 0=禁用 1=启用"
// @Success 200 {object} dto.Response{data=dto.PageData{items=[]admindto.RoleItem}}
// @Failure 401 {object} dto.Response
// @Router /admin/system/role/list [get]
func GetRoleList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	name := c.Query("name")
	code := c.Query("code")
	status := c.Query("status")

	var roles []adminmodel.Role
	var total int64
	query := store.DB.Model(&adminmodel.Role{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if status == "0" || status == "1" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	query.Preload("Menus").Offset(offset).Limit(pageSize).Find(&roles)

	items := make([]admindto.RoleItem, len(roles))
	for i, r := range roles {
		menuIDs := make([]uint, len(r.Menus))
		for j, m := range r.Menus {
			menuIDs[j] = m.ID
		}
		items[i] = admindto.RoleItem{
			ID:          r.ID,
			Name:        r.Name,
			Code:        r.Code,
			Status:      r.Status,
			Remark:      r.Remark,
			Permissions: menuIDs,
			CreateTime:  r.CreatedAt.Format("2006/01/02 15:04:05"),
		}
	}

	return dto.PageSuccess(c, items, total)
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建新角色并分配菜单权限
// @Tags 系统管理 - 角色
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body admindto.CreateRoleRequest true "角色信息"
// @Success 200 {object} dto.Response{data=dto.IDResponse}
// @Failure 400 {object} dto.Response
// @Router /admin/system/role [post]
func CreateRole(c *fiber.Ctx) error {
	var req admindto.CreateRoleRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	role := adminmodel.Role{
		Name:   req.Name,
		Code:   req.Code,
		Status: req.Status,
		Remark: req.Remark,
	}
	if err := store.DB.Create(&role).Error; err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to create role: "+err.Error())
	}

	if len(req.MenuIDs) > 0 {
		var menus []adminmodel.Menu
		store.DB.Where("id IN ?", req.MenuIDs).Find(&menus)
		store.DB.Model(&role).Association("Menus").Replace(menus)
	}

	return dto.Success(c, fiber.Map{"id": role.ID})
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息和菜单权限
// @Tags 系统管理 - 角色
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Param request body admindto.CreateRoleRequest true "角色信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /admin/system/role/{id} [put]
func UpdateRole(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var req admindto.CreateRoleRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	updates := map[string]any{
		"name":   req.Name,
		"code":   req.Code,
		"status": req.Status,
		"remark": req.Remark,
	}
	store.DB.Model(&adminmodel.Role{}).Where("id = ?", id).Updates(updates)

	if req.MenuIDs != nil {
		var role adminmodel.Role
		store.DB.First(&role, id)
		var menus []adminmodel.Menu
		store.DB.Where("id IN ?", req.MenuIDs).Find(&menus)
		store.DB.Model(&role).Association("Menus").Replace(menus)
	}

	return dto.Success(c, nil)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除指定角色及其权限关联
// @Tags 系统管理 - 角色
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Success 200 {object} dto.Response
// @Router /admin/system/role/{id} [delete]
func DeleteRole(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var role adminmodel.Role
	if err := store.DB.First(&role, id).Error; err != nil {
		return dto.Fail(c, fiber.StatusNotFound, "Role not found")
	}
	store.DB.Model(&role).Association("Menus").Clear()
	store.DB.Delete(&role)
	return dto.Success(c, nil)
}
