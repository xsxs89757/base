package admin

import v "base/internal/validator"

func init() {
	v.RegisterMessage("DeptRequest.Name.required", "部门名称不能为空")
	v.RegisterMessage("DeptRequest.Status.oneof", "状态取值不合法")
}
