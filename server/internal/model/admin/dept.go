package admin

import "base/internal/model"

type Dept struct {
	model.BaseModel
	ParentID uint   `json:"pid" gorm:"default:0"`
	Name     string `json:"name" gorm:"size:64;not null"`
	OrderNo  int    `json:"order" gorm:"default:0"`
	Status   int    `json:"status" gorm:"default:1"`
	Remark   string `json:"remark" gorm:"size:256"`
}

func (Dept) TableName() string {
	return "sys_depts"
}
