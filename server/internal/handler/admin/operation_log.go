package admin

import (
	"strconv"

	"base/internal/dto"
	admindto "base/internal/dto/admin"
	adminmodel "base/internal/model/admin"
	"base/internal/store"

	"github.com/gofiber/fiber/v2"
)

// GetOperationLogList 获取操作日志列表
// @Summary 获取操作日志列表
// @Description 分页查询操作日志，支持按用户名、请求方法、路径和状态筛选
// @Tags 系统管理 - 操作日志
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param username query string false "操作用户(模糊搜索)"
// @Param method query string false "请求方法: GET/POST/PUT/DELETE"
// @Param path query string false "请求路径(模糊搜索)"
// @Param status query string false "响应状态码"
// @Success 200 {object} dto.Response{data=dto.PageData{items=[]admindto.OperationLogItem}}
// @Router /admin/system/operation-log/list [get]
func GetOperationLogList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	username := c.Query("username")
	method := c.Query("method")
	path := c.Query("path")
	status := c.Query("status")

	var logs []adminmodel.OperationLog
	var total int64
	query := store.DB.Model(&adminmodel.OperationLog{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if method != "" {
		query = query.Where("method = ?", method)
	}
	if path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs)

	items := make([]admindto.OperationLogItem, len(logs))
	for i, l := range logs {
		items[i] = admindto.OperationLogItem{
			ID:         l.ID,
			Username:   l.Username,
			Method:     l.Method,
			Path:       l.Path,
			Status:     l.Status,
			Duration:   l.Duration,
			IP:         l.IP,
			UserAgent:  l.UserAgent,
			CreateTime: l.CreatedAt.Format("2006/01/02 15:04:05"),
		}
	}
	return dto.PageSuccess(c, items, total)
}

// DeleteOperationLog 删除操作日志
// @Summary 删除操作日志
// @Tags 系统管理 - 操作日志
// @Produce json
// @Security BearerAuth
// @Param id path int true "日志ID"
// @Success 200 {object} dto.Response
// @Router /admin/system/operation-log/{id} [delete]
func DeleteOperationLog(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	store.DB.Delete(&adminmodel.OperationLog{}, id)
	return dto.Success(c, nil)
}

// ClearOperationLog 清空操作日志
// @Summary 清空操作日志
// @Tags 系统管理 - 操作日志
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /admin/system/operation-log/clear [delete]
func ClearOperationLog(c *fiber.Ctx) error {
	store.DB.Where("1 = 1").Delete(&adminmodel.OperationLog{})
	return dto.Success(c, nil)
}
