package admin

import (
	"errors"

	adminmodel "base/internal/model/admin"
	"base/internal/store"

	"golang.org/x/crypto/bcrypt"
)

var ErrSuperAdminProtected = errors.New("超级管理员不可修改或删除")

type UserListParams struct {
	Page     int
	PageSize int
	Username string
	Status   *int
}

func GetUserList(params UserListParams) ([]adminmodel.User, int64, error) {
	var users []adminmodel.User
	var total int64

	query := store.DB.Model(&adminmodel.User{}).Preload("Roles").Where("id != ?", 1)

	if params.Username != "" {
		query = query.Where("username LIKE ?", "%"+params.Username+"%")
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	query.Count(&total)

	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func CreateUser(user *adminmodel.User, roleIDs []uint) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)

	if err := store.DB.Create(user).Error; err != nil {
		return err
	}
	if len(roleIDs) > 0 {
		var roles []adminmodel.Role
		store.DB.Where("id IN ?", roleIDs).Find(&roles)
		return store.DB.Model(user).Association("Roles").Replace(roles)
	}
	return nil
}

func UpdateUser(id uint, updates map[string]any, roleIDs []uint) error {
	if id == 1 {
		return ErrSuperAdminProtected
	}
	if pwd, ok := updates["password"]; ok && pwd != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(pwd.(string)), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password"] = string(hash)
	} else {
		delete(updates, "password")
	}

	if err := store.DB.Model(&adminmodel.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	if roleIDs != nil {
		var user adminmodel.User
		store.DB.First(&user, id)
		var roles []adminmodel.Role
		store.DB.Where("id IN ?", roleIDs).Find(&roles)
		return store.DB.Model(&user).Association("Roles").Replace(roles)
	}
	return nil
}

func NewUser(username, password, realName, email, phone string, status int, remark string) *adminmodel.User {
	return &adminmodel.User{
		Username: username,
		Password: password,
		RealName: realName,
		Email:    email,
		Phone:    phone,
		Status:   status,
		Remark:   remark,
	}
}

func DeleteUser(id uint) error {
	if id == 1 {
		return ErrSuperAdminProtected
	}
	var user adminmodel.User
	if err := store.DB.First(&user, id).Error; err != nil {
		return err
	}
	store.DB.Model(&user).Association("Roles").Clear()
	return store.DB.Delete(&user).Error
}
