package admin

import (
	adminmodel "base/internal/model/admin"
	"base/internal/store"

	"golang.org/x/crypto/bcrypt"
)

func Authenticate(username, password string) (*adminmodel.User, error) {
	var user adminmodel.User
	if err := store.DB.Preload("Roles").Where("username = ? AND status = 1", username).First(&user).Error; err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id uint) (*adminmodel.User, error) {
	var user adminmodel.User
	if err := store.DB.Preload("Roles").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByUsername(username string) (*adminmodel.User, error) {
	var user adminmodel.User
	if err := store.DB.Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func VerifyPassword(user *adminmodel.User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}

func ChangePassword(userID uint, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return store.DB.Model(&adminmodel.User{}).Where("id = ?", userID).Update("password", string(hash)).Error
}

func GetRoleNames(user *adminmodel.User) []string {
	names := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		names[i] = r.Code
	}
	return names
}

func GetAccessCodes(user *adminmodel.User) []string {
	var codes []string
	var menus []adminmodel.Menu

	if user.ID == 1 {
		store.DB.Where("auth_code != ''").Find(&menus)
	} else {
		var roleIDs []uint
		for _, r := range user.Roles {
			roleIDs = append(roleIDs, r.ID)
		}
		store.DB.
			Joins("JOIN role_menus ON role_menus.menu_id = sys_menus.id").
			Where("role_menus.role_id IN ? AND sys_menus.auth_code != ''", roleIDs).
			Find(&menus)
	}

	seen := make(map[string]bool)
	for _, m := range menus {
		if m.AuthCode != "" && !seen[m.AuthCode] {
			codes = append(codes, m.AuthCode)
			seen[m.AuthCode] = true
		}
	}
	return codes
}
