package dto

type SetStatusRequest struct {
	Status string `json:"status" validate:"required" example:"active"`
}

type ApprovePioneerRequest struct {
	Note string `json:"note,omitempty" example:"ok"`
}

type RejectPioneerRequest struct {
	Reason string `json:"reason" validate:"required" example:"invalid docs"`
}
