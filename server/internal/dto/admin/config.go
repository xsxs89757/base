package admin

// ConfigRequest 创建/更新配置请求
type ConfigRequest struct {
	ConfigKey   string `json:"configKey" validate:"required" example:"site_name"`
	ConfigValue string `json:"configValue" example:"Admin"`
	ConfigGroup string `json:"configGroup" example:"basic"`
	Remark      string `json:"remark"`
	Status      int    `json:"status" validate:"oneof=0 1" example:"1"`
}

// ConfigItem 配置列表项
type ConfigItem struct {
	ID          uint   `json:"id" example:"1"`
	ConfigKey   string `json:"configKey" example:"site_name"`
	ConfigValue string `json:"configValue" example:"Admin"`
	ConfigGroup string `json:"configGroup" example:"basic"`
	Remark      string `json:"remark"`
	Status      int    `json:"status" example:"1"`
	CreateTime  string `json:"createTime"`
}
