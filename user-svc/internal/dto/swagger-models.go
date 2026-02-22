package dto

// ===== Common responses =====

type APIError struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"invalid input"`
}

type APISuccessString struct {
	Success bool   `json:"success" example:"true"`
	Data    string `json:"data" example:"ok"`
}

type APISuccessAny struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data"`
}

type APISuccessLogin struct {
	Success bool          `json:"success" example:"true"`
	Data    LoginResponse `json:"data"`
}
