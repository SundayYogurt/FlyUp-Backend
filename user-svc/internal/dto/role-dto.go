package dto

type SetRolesRequest struct {
	Roles []string `json:"roles" validate:"required,min=1"` // ["BOOSTER","PIONEER"]
}

type RoleResponse struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type UserRolesResponse struct {
	UserID uint           `json:"user_id"`
	Roles  []RoleResponse `json:"roles"`
}
