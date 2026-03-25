package admin

// MenuRequest 创建/更新菜单请求
type MenuRequest struct {
	ParentID  uint   `json:"pid" example:"0"`
	Name      string `json:"name" validate:"required" example:"Dashboard"`
	Path      string `json:"path" example:"/dashboard"`
	Component string `json:"component" example:"/dashboard/index"`
	Redirect  string `json:"redirect"`
	Type      string `json:"type" validate:"required,oneof=catalog menu button embedded link" example:"menu"`
	Icon      string `json:"icon" example:"lucide:layout-dashboard"`
	Title     string `json:"title" validate:"required" example:"仪表盘"`
	AuthCode  string `json:"authCode" example:"System:Menu:List"`
	OrderNo   int    `json:"order" example:"1"`
	Status    int    `json:"status" validate:"oneof=0 1" example:"1"`
	KeepAlive bool   `json:"keepAlive"`
	AffixTab  bool   `json:"affixTab"`
	IframeSrc string `json:"iframeSrc"`
	Link      string `json:"link"`
}
