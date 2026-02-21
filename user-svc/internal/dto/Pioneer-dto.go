package dto

type PioneerInput struct {
	StudentCode    string  `json:"student_code" validate:"required"`
	Faculty        string  `json:"faculty" validate:"required"`
	Major          string  `json:"major" validate:"required"`
	YearLevel      string  `json:"year_level" validate:"required"`
	StudentCardURL *string `json:"student_card_url" validate:"required,uri"`
}

type PioneerResponse struct {
	UserID         uint    `json:"user_id"`
	UniversityID   *uint   `json:"university_id,omitempty"`
	StudentCode    string  `json:"student_code"`
	Faculty        string  `json:"faculty"`
	Major          string  `json:"major"`
	YearLevel      string  `json:"year_level"`
	StudentCardURL *string `json:"student_card_url,omitempty"`
	VerifyStatus   string  `json:"verify_status"` // pending | approved | rejected
	Note           *string `json:"note,omitempty"`
}
