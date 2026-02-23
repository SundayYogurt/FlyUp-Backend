package dto

type PioneerInput struct {
	StudentCode    string   `json:"student_code" validate:"required"`
	Faculty        string   `json:"faculty" validate:"required"`
	Major          string   `json:"major" validate:"required"`
	StudentCardURL *string  `json:"student_card_url" validate:"required,uri"`
	ConsentCodes   []string `json:"consent_codes"`
}

type PioneerResponse struct {
	UserID         uint    `json:"user_id"`
	UniversityID   *uint   `json:"university_id,omitempty"`
	StudentCode    string  `json:"student_code"`
	Faculty        string  `json:"faculty"`
	Major          string  `json:"major"`
	StudentCardURL *string `json:"student_card_url,omitempty"`
	VerifyStatus   string  `json:"verify_status"` // pending | approved | rejected
	Note           *string `json:"note,omitempty"`
}
