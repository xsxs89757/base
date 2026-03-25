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

// GetDeptList 获取部门列表
// @Summary 获取部门树形列表
// @Description 返回所有部门的树形结构
// @Tags 系统管理 - 部门
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]object}
// @Failure 401 {object} dto.Response
// @Router /admin/system/dept/list [get]
func GetDeptList(c *fiber.Ctx) error {
	var depts []adminmodel.Dept
	store.DB.Order("order_no ASC").Find(&depts)
	tree := buildDeptTree(depts, 0)
	return dto.Success(c, tree)
}

// CreateDept 创建部门
// @Summary 创建部门
// @Description 创建新部门
// @Tags 系统管理 - 部门
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body admindto.DeptRequest true "部门信息"
// @Success 200 {object} dto.Response{data=dto.IDResponse}
// @Failure 400 {object} dto.Response
// @Router /admin/system/dept [post]
func CreateDept(c *fiber.Ctx) error {
	var req admindto.DeptRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	dept := adminmodel.Dept{
		ParentID: req.ParentID,
		Name:     req.Name,
		OrderNo:  req.OrderNo,
		Status:   req.Status,
		Remark:   req.Remark,
	}
	if err := store.DB.Create(&dept).Error; err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to create dept")
	}
	return dto.Success(c, fiber.Map{"id": dept.ID})
}

// UpdateDept 更新部门
// @Summary 更新部门
// @Description 更新部门信息
// @Tags 系统管理 - 部门
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param request body admindto.DeptRequest true "部门信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /admin/system/dept/{id} [put]
func UpdateDept(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var req admindto.DeptRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	updates := map[string]any{
		"parent_id": req.ParentID,
		"name":      req.Name,
		"order_no":  req.OrderNo,
		"status":    req.Status,
		"remark":    req.Remark,
	}
	store.DB.Model(&adminmodel.Dept{}).Where("id = ?", id).Updates(updates)
	return dto.Success(c, nil)
}

// DeleteDept 删除部门
// @Summary 删除部门
// @Description 删除指定部门及其子部门
// @Tags 系统管理 - 部门
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Success 200 {object} dto.Response
// @Router /admin/system/dept/{id} [delete]
func DeleteDept(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	store.DB.Where("parent_id = ?", id).Delete(&adminmodel.Dept{})
	store.DB.Delete(&adminmodel.Dept{}, id)
	return dto.Success(c, nil)
}

func buildDeptTree(depts []adminmodel.Dept, parentID uint) []fiber.Map {
	var tree []fiber.Map
	for _, d := range depts {
		if d.ParentID == parentID {
			node := fiber.Map{
				"id":         d.ID,
				"pid":        d.ParentID,
				"name":       d.Name,
				"status":     d.Status,
				"remark":     d.Remark,
				"order":      d.OrderNo,
				"createTime": d.CreatedAt.Format("2006/01/02 15:04:05"),
			}
			children := buildDeptTree(depts, d.ID)
			if len(children) > 0 {
				node["children"] = children
			}
			tree = append(tree, node)
		}
	}
	return tree
}
