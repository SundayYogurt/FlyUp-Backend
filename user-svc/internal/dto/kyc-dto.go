package dto

type KYCRequest struct {
	Documents []KYCDocumentInput `json:"documents"`
}

type KYCDocumentInput struct {
	DocType string `json:"doc_type"` // id_card | student_card | selfie | other
	FileURL string `json:"file_url"`
}

type KYCStatusResponse struct {
	KYCID       uint                  `json:"kyc_id"`
	Status      string                `json:"status"` // pending | approved | rejected
	SubmittedAt string                `json:"submitted_at"`
	ReviewedAt  *string               `json:"reviewed_at,omitempty"`
	ReviewNote  *string               `json:"review_note,omitempty"`
	Documents   []KYCDocumentResponse `json:"documents"`
}

type KYCDocumentResponse struct {
	ID      uint   `json:"id"`
	DocType string `json:"doc_type"`
	FileURL string `json:"file_url"`
}

type PendingKYCResponse struct {
	KYCID       uint   `json:"kyc_id"`
	UserID      uint   `json:"user_id"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submitted_at"`
}

type ApproveKYCRequest struct {
	Note string `json:"note"`
}

type RejectKYCRequest struct {
	Reason string `json:"reason"`
}
