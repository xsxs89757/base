package admin

import "base/internal/model"

type Config struct {
	model.BaseModel
	ConfigKey   string `json:"configKey" gorm:"uniqueIndex;not null;size:128"`
	ConfigValue string `json:"configValue" gorm:"type:text"`
	ConfigGroup string `json:"configGroup" gorm:"size:64;index"`
	Remark      string `json:"remark" gorm:"size:256"`
	Status      int    `json:"status" gorm:"default:1"`
}

func (Config) TableName() string {
	return "sys_configs"
}
