package admin

import (
	"sort"
	"strconv"

	"base/internal/dto"
	admindto "base/internal/dto/admin"
	adminmodel "base/internal/model/admin"
	adminsvc "base/internal/service/admin"
	"base/internal/store"
	"base/internal/validator"

	"github.com/gofiber/fiber/v2"
)

// GetAllMenus 获取用户菜单
// @Summary 获取当前用户的菜单树
// @Description 根据当前用户角色返回可访问的菜单树(用于前端动态路由)
// @Tags 菜单
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]object}
// @Failure 401 {object} dto.Response
// @Router /admin/menu/all [get]
func GetAllMenus(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	user, err := adminsvc.GetUserByUsername(username)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "User not found")
	}

	var menus []adminmodel.Menu

	if user.ID == 1 {
		store.DB.
			Where("type != ? AND status = ?", "button", 1).
			Order("order_no ASC").
			Find(&menus)
	} else {
		var roleIDs []uint
		for _, r := range user.Roles {
			roleIDs = append(roleIDs, r.ID)
		}
		store.DB.
			Joins("JOIN role_menus ON role_menus.menu_id = sys_menus.id").
			Where("role_menus.role_id IN ? AND sys_menus.type IN ? AND sys_menus.status = ?", roleIDs, []string{"menu", "embedded", "link"}, 1).
			Order("sys_menus.order_no ASC").
			Distinct().
			Find(&menus)
		menus = includeMenuAncestors(menus)
	}

	tree := buildMenuTree(menus, 0)
	return dto.Success(c, tree)
}

// GetMenuList 获取菜单管理列表
// @Summary 获取菜单管理列表
// @Description 返回所有菜单的树形结构(用于后台菜单管理页面)
// @Tags 系统管理 - 菜单
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]object}
// @Failure 401 {object} dto.Response
// @Router /admin/system/menu/list [get]
func GetMenuList(c *fiber.Ctx) error {
	var menus []adminmodel.Menu
	store.DB.Order("order_no ASC").Find(&menus)
	tree := buildMenuTreeForManage(menus, 0)
	return dto.Success(c, tree)
}

// CreateMenu 创建菜单
// @Summary 创建菜单
// @Description 创建新的菜单/目录/按钮
// @Tags 系统管理 - 菜单
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body admindto.MenuRequest true "菜单信息"
// @Success 200 {object} dto.Response{data=dto.IDResponse}
// @Failure 400 {object} dto.Response
// @Router /admin/system/menu [post]
func CreateMenu(c *fiber.Ctx) error {
	var req admindto.MenuRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	menu := adminmodel.Menu{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Redirect:  req.Redirect,
		Type:      req.Type,
		Icon:      req.Icon,
		Title:     req.Title,
		AuthCode:  req.AuthCode,
		OrderNo:   req.OrderNo,
		Status:    req.Status,
		KeepAlive: req.KeepAlive,
		AffixTab:  req.AffixTab,
		IframeSrc: req.IframeSrc,
		Link:      req.Link,
	}
	if err := store.DB.Create(&menu).Error; err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to create menu")
	}
	return dto.Success(c, fiber.Map{"id": menu.ID})
}

// UpdateMenu 更新菜单
// @Summary 更新菜单
// @Description 更新菜单信息
// @Tags 系统管理 - 菜单
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "菜单ID"
// @Param request body admindto.MenuRequest true "菜单信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /admin/system/menu/{id} [put]
func UpdateMenu(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var req admindto.MenuRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	updates := adminmodel.Menu{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Redirect:  req.Redirect,
		Type:      req.Type,
		Icon:      req.Icon,
		Title:     req.Title,
		AuthCode:  req.AuthCode,
		OrderNo:   req.OrderNo,
		Status:    req.Status,
		KeepAlive: req.KeepAlive,
		AffixTab:  req.AffixTab,
		IframeSrc: req.IframeSrc,
		Link:      req.Link,
	}
	store.DB.Model(&adminmodel.Menu{}).Where("id = ?", id).Updates(updates)
	return dto.Success(c, nil)
}

// DeleteMenu 删除菜单
// @Summary 删除菜单
// @Description 删除指定菜单及其子菜单
// @Tags 系统管理 - 菜单
// @Produce json
// @Security BearerAuth
// @Param id path int true "菜单ID"
// @Success 200 {object} dto.Response
// @Router /admin/system/menu/{id} [delete]
func DeleteMenu(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	store.DB.Where("parent_id = ?", id).Delete(&adminmodel.Menu{})
	store.DB.Delete(&adminmodel.Menu{}, id)
	return dto.Success(c, nil)
}

// CheckMenuNameExists 检查菜单名称是否存在
// @Summary 检查菜单名称是否存在
// @Tags 系统管理 - 菜单
// @Produce json
// @Security BearerAuth
// @Param name query string true "菜单名称"
// @Success 200 {object} dto.Response{data=bool}
// @Router /admin/system/menu/name-exists [get]
func CheckMenuNameExists(c *fiber.Ctx) error {
	name := c.Query("name")
	var count int64
	store.DB.Model(&adminmodel.Menu{}).Where("name = ?", name).Count(&count)
	return dto.Success(c, count > 0)
}

// CheckMenuPathExists 检查菜单路径是否存在
// @Summary 检查菜单路径是否存在
// @Tags 系统管理 - 菜单
// @Produce json
// @Security BearerAuth
// @Param path query string true "路由路径"
// @Success 200 {object} dto.Response{data=bool}
// @Router /admin/system/menu/path-exists [get]
func CheckMenuPathExists(c *fiber.Ctx) error {
	path := c.Query("path")
	var count int64
	store.DB.Model(&adminmodel.Menu{}).Where("path = ?", path).Count(&count)
	return dto.Success(c, count > 0)
}

func buildMenuTree(menus []adminmodel.Menu, parentID uint) []fiber.Map {
	var tree []fiber.Map
	for _, m := range menus {
		if m.ParentID == parentID {
			node := fiber.Map{
				"name": m.Name,
				"path": m.Path,
				"meta": fiber.Map{
					"title":     m.Title,
					"icon":      m.Icon,
					"order":     m.OrderNo,
					"affixTab":  m.AffixTab,
					"keepAlive": m.KeepAlive,
				},
			}
			if m.Component != "" {
				node["component"] = m.Component
			}
			if m.Redirect != "" {
				node["redirect"] = m.Redirect
			}
			children := buildMenuTree(menus, m.ID)
			if len(children) > 0 {
				node["children"] = children
			}
			tree = append(tree, node)
		}
	}
	return tree
}

func includeMenuAncestors(menus []adminmodel.Menu) []adminmodel.Menu {
	if len(menus) == 0 {
		return menus
	}

	var allMenus []adminmodel.Menu
	store.DB.Where("type != ? AND status = ?", "button", 1).Find(&allMenus)

	allByID := make(map[uint]adminmodel.Menu, len(allMenus))
	for _, menu := range allMenus {
		allByID[menu.ID] = menu
	}

	selectedByID := make(map[uint]adminmodel.Menu, len(menus))
	for _, menu := range menus {
		selectedByID[menu.ID] = menu
	}

	for _, menu := range menus {
		visited := map[uint]bool{}
		for parentID := menu.ParentID; parentID != 0 && !visited[parentID]; {
			visited[parentID] = true
			parent, ok := allByID[parentID]
			if !ok {
				break
			}
			if _, exists := selectedByID[parent.ID]; !exists {
				selectedByID[parent.ID] = parent
			}
			parentID = parent.ParentID
		}
	}

	result := make([]adminmodel.Menu, 0, len(selectedByID))
	for _, menu := range selectedByID {
		result = append(result, menu)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].OrderNo == result[j].OrderNo {
			return result[i].ID < result[j].ID
		}
		return result[i].OrderNo < result[j].OrderNo
	})
	return result
}

func buildMenuTreeForManage(menus []adminmodel.Menu, parentID uint) []fiber.Map {
	var tree []fiber.Map
	for _, m := range menus {
		if m.ParentID == parentID {
			node := fiber.Map{
				"id":     m.ID,
				"pid":    m.ParentID,
				"name":   m.Name,
				"path":   m.Path,
				"status": m.Status,
				"type":   m.Type,
				"meta": fiber.Map{
					"title": m.Title,
					"icon":  m.Icon,
					"order": m.OrderNo,
				},
			}
			if m.Component != "" {
				node["component"] = m.Component
			}
			if m.AuthCode != "" {
				node["authCode"] = m.AuthCode
			}
			children := buildMenuTreeForManage(menus, m.ID)
			if len(children) > 0 {
				node["children"] = children
			}
			tree = append(tree, node)
		}
	}
	return tree
}
