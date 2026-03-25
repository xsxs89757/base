package admin

import "base/internal/model"

type Menu struct {
	model.BaseModel
	ParentID  uint   `json:"pid" gorm:"default:0;comment:parent menu id"`
	Name      string `json:"name" gorm:"size:64;not null"`
	Path      string `json:"path" gorm:"size:128"`
	Component string `json:"component" gorm:"size:128"`
	Redirect  string `json:"redirect,omitempty" gorm:"size:128"`
	Type      string `json:"type" gorm:"size:16;not null;comment:catalog|menu|button|embedded|link"`
	Icon      string `json:"icon" gorm:"size:64"`
	Title     string `json:"title" gorm:"size:64;not null"`
	AuthCode  string `json:"authCode,omitempty" gorm:"size:64"`
	OrderNo   int    `json:"order" gorm:"default:0"`
	Status    int    `json:"status" gorm:"default:1;comment:0=disabled 1=enabled"`
	KeepAlive bool   `json:"keepAlive" gorm:"default:false"`
	AffixTab  bool   `json:"affixTab" gorm:"default:false"`
	IframeSrc string `json:"iframeSrc,omitempty" gorm:"size:256"`
	Link      string `json:"link,omitempty" gorm:"size:256"`
	Children  []Menu `json:"children,omitempty" gorm:"-"`
}

func (Menu) TableName() string {
	return "sys_menus"
}
