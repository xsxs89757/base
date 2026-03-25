package admin

import (
	"time"

	"base/config"
	"base/internal/dto"
	admindto "base/internal/dto/admin"
	"base/internal/middleware"
	adminsvc "base/internal/service/admin"
	"base/internal/validator"

	"github.com/gofiber/fiber/v2"
)

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名和密码登录，返回 accessToken
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body admindto.LoginRequest true "登录参数"
// @Success 200 {object} dto.Response{data=admindto.LoginResponse}
// @Failure 400 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Router /admin/auth/login [post]
func Login(c *fiber.Ctx) error {
	var req admindto.LoginRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := adminsvc.Authenticate(req.Username, req.Password)
	if err != nil {
		return dto.Fail(c, fiber.StatusForbidden, "Username or password is incorrect.")
	}

	roles := adminsvc.GetRoleNames(user)
	accessToken, err := middleware.GenerateAccessToken(user.ID, user.Username, roles)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to generate token")
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to generate refresh token")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    refreshToken,
		HTTPOnly: true,
		SameSite: "None",
		Secure:   true,
		MaxAge:   int(config.C.JWT.RefreshExpire / time.Second),
	})

	return dto.Success(c, fiber.Map{
		"id":          user.ID,
		"username":    user.Username,
		"realName":    user.RealName,
		"avatar":      user.Avatar,
		"roles":       roles,
		"homePath":    user.HomePath,
		"accessToken": accessToken,
	})
}

// Logout 退出登录
// @Summary 退出登录
// @Description 清除 refresh token cookie
// @Tags 认证
// @Produce json
// @Success 200 {object} dto.Response
// @Router /admin/auth/logout [post]
func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		HTTPOnly: true,
		SameSite: "None",
		Secure:   true,
		MaxAge:   -1,
	})
	return dto.Success(c, "")
}

// RefreshToken 刷新令牌
// @Summary 刷新 Access Token
// @Description 使用 cookie 中的 refresh token 获取新的 access token
// @Tags 认证
// @Produce plain
// @Success 200 {string} string "新的 access token"
// @Failure 403 {object} dto.Response
// @Router /admin/auth/refresh [post]
func RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("jwt")
	if refreshToken == "" {
		return dto.Fail(c, fiber.StatusForbidden, "Forbidden Exception")
	}

	claims, err := middleware.ParseToken(refreshToken)
	if err != nil {
		c.Cookie(&fiber.Cookie{Name: "jwt", Value: "", MaxAge: -1})
		return dto.Fail(c, fiber.StatusForbidden, "Forbidden Exception")
	}

	user, err := adminsvc.GetUserByUsername(claims.Username)
	if err != nil {
		return dto.Fail(c, fiber.StatusForbidden, "Forbidden Exception")
	}

	roles := adminsvc.GetRoleNames(user)
	newToken, err := middleware.GenerateAccessToken(user.ID, user.Username, roles)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to generate token")
	}

	return c.SendString(newToken)
}

// ChangePassword 修改当前用户密码
// @Summary 修改密码
// @Description 用户修改自己的密码，需验证旧密码
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body admindto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Router /admin/auth/change-password [post]
func ChangePassword(c *fiber.Ctx) error {
	var req admindto.ChangePasswordRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	userID := c.Locals("userId").(uint)
	user, err := adminsvc.GetUserByID(userID)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "用户不存在")
	}

	if err := adminsvc.VerifyPassword(user, req.OldPassword); err != nil {
		return dto.Fail(c, fiber.StatusForbidden, "旧密码不正确")
	}

	if err := adminsvc.ChangePassword(userID, req.NewPassword); err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "修改密码失败")
	}

	return dto.Success(c, nil)
}

// GetAccessCodes 获取权限码
// @Summary 获取当前用户权限码
// @Description 返回当前登录用户拥有的所有权限码列表
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]string}
// @Failure 401 {object} dto.Response
// @Router /admin/auth/codes [get]
func GetAccessCodes(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	user, err := adminsvc.GetUserByUsername(username)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "User not found")
	}
	codes := adminsvc.GetAccessCodes(user)
	if codes == nil {
		codes = []string{}
	}
	return dto.Success(c, codes)
}

// GetUserInfo 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 返回当前登录用户的基本信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=admindto.UserInfoResponse}
// @Failure 401 {object} dto.Response
// @Router /admin/user/info [get]
func GetUserInfo(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uint)
	user, err := adminsvc.GetUserByID(userID)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "User not found")
	}
	roles := adminsvc.GetRoleNames(user)
	return dto.Success(c, fiber.Map{
		"id":       user.ID,
		"username": user.Username,
		"realName": user.RealName,
		"avatar":   user.Avatar,
		"roles":    roles,
		"homePath": user.HomePath,
	})
}
