package dto

import "github.com/gofiber/fiber/v2"

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code" example:"0"`
	Data    any    `json:"data"`
	Error   any    `json:"error"`
	Message string `json:"message" example:"ok"`
}

// PageData 分页数据
type PageData struct {
	Items any   `json:"items"`
	Total int64 `json:"total" example:"100"`
}

// PageResponse 分页响应 (Swagger 用)
type PageResponse struct {
	Code    int      `json:"code" example:"0"`
	Data    PageData `json:"data"`
	Error   any      `json:"error"`
	Message string   `json:"message" example:"ok"`
}

// IDResponse 创建成功返回 ID
type IDResponse struct {
	ID uint `json:"id" example:"1"`
}

func Success(c *fiber.Ctx, data any) error {
	return c.JSON(Response{
		Code:    0,
		Data:    data,
		Error:   nil,
		Message: "ok",
	})
}

func PageSuccess(c *fiber.Ctx, items any, total int64) error {
	return c.JSON(Response{
		Code: 0,
		Data: PageData{
			Items: items,
			Total: total,
		},
		Error:   nil,
		Message: "ok",
	})
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Code:    -1,
		Data:    nil,
		Error:   message,
		Message: message,
	})
}
