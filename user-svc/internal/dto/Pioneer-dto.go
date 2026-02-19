package dto

type PioneerVerificationRequest struct {
	StudentCode    string `json:"student_code" validate:"required"`
	Faculty        string `json:"faculty" validate:"required"`
	Major          string `json:"major" validate:"required"`
	YearLevel      string `json:"year_level" validate:"required"`
	StudentCardURL string `json:"student_card_url" validate:"required,url"`
}

type PioneerVerificationResponse struct {
	UserID         uint   `json:"user_id"`
	UniversityID   uint   `json:"university_id"`
	StudentCode    string `json:"student_code"`
	Faculty        string `json:"faculty"`
	Major          string `json:"major"`
	YearLevel      string `json:"year_level"`
	StudentCardURL string `json:"student_card_url"`
	VerifyStatus   string `json:"verify_status"`
	Note           string `json:"note,omitempty"`
}
