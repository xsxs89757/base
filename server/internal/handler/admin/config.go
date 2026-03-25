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

// GetConfigList 获取配置列表
// @Summary 获取配置列表
// @Description 分页查询配置列表，支持按 key、分组和状态筛选
// @Tags 系统管理 - 配置
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param configKey query string false "配置键(模糊搜索)"
// @Param configGroup query string false "配置分组"
// @Param status query string false "状态: 0=禁用 1=启用"
// @Success 200 {object} dto.Response{data=dto.PageData{items=[]admindto.ConfigItem}}
// @Router /admin/system/config/list [get]
func GetConfigList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	configKey := c.Query("configKey")
	configGroup := c.Query("configGroup")
	status := c.Query("status")

	var configs []adminmodel.Config
	var total int64
	query := store.DB.Model(&adminmodel.Config{})

	if configKey != "" {
		query = query.Where("config_key LIKE ?", "%"+configKey+"%")
	}
	if configGroup != "" {
		query = query.Where("config_group = ?", configGroup)
	}
	if status == "0" || status == "1" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&configs)

	items := make([]admindto.ConfigItem, len(configs))
	for i, cfg := range configs {
		items[i] = admindto.ConfigItem{
			ID:          cfg.ID,
			ConfigKey:   cfg.ConfigKey,
			ConfigValue: cfg.ConfigValue,
			ConfigGroup: cfg.ConfigGroup,
			Remark:      cfg.Remark,
			Status:      cfg.Status,
			CreateTime:  cfg.CreatedAt.Format("2006/01/02 15:04:05"),
		}
	}
	return dto.PageSuccess(c, items, total)
}

// GetConfigGroups 获取配置分组列表
// @Summary 获取配置分组列表
// @Description 返回所有不重复的配置分组名称
// @Tags 系统管理 - 配置
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]string}
// @Router /admin/system/config/groups [get]
func GetConfigGroups(c *fiber.Ctx) error {
	var groups []string
	store.DB.Model(&adminmodel.Config{}).
		Where("config_group != ''").
		Distinct("config_group").
		Order("config_group").
		Pluck("config_group", &groups)
	return dto.Success(c, groups)
}

// CreateConfig 创建配置
// @Summary 创建配置
// @Tags 系统管理 - 配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body admindto.ConfigRequest true "配置信息"
// @Success 200 {object} dto.Response{data=dto.IDResponse}
// @Router /admin/system/config [post]
func CreateConfig(c *fiber.Ctx) error {
	var req admindto.ConfigRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	cfg := adminmodel.Config{
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
		ConfigGroup: req.ConfigGroup,
		Remark:      req.Remark,
		Status:      req.Status,
	}
	if err := store.DB.Create(&cfg).Error; err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to create config: "+err.Error())
	}
	return dto.Success(c, fiber.Map{"id": cfg.ID})
}

// UpdateConfig 更新配置
// @Summary 更新配置
// @Tags 系统管理 - 配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "配置ID"
// @Param request body admindto.ConfigRequest true "配置信息"
// @Success 200 {object} dto.Response
// @Router /admin/system/config/{id} [put]
func UpdateConfig(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var req admindto.ConfigRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	updates := map[string]any{
		"config_key":   req.ConfigKey,
		"config_value": req.ConfigValue,
		"config_group": req.ConfigGroup,
		"remark":       req.Remark,
		"status":       req.Status,
	}
	if err := store.DB.Model(&adminmodel.Config{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to update config")
	}
	return dto.Success(c, nil)
}

// DeleteConfig 删除配置
// @Summary 删除配置
// @Tags 系统管理 - 配置
// @Produce json
// @Security BearerAuth
// @Param id path int true "配置ID"
// @Success 200 {object} dto.Response
// @Router /admin/system/config/{id} [delete]
func DeleteConfig(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := store.DB.Delete(&adminmodel.Config{}, id).Error; err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to delete config")
	}
	return dto.Success(c, nil)
}
