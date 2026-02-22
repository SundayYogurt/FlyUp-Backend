package dto

type KYCSubmitResponse struct {
	KYCID        uint    `json:"kyc_id"`
	Status       string  `json:"status"`
	OCRProvider  *string `json:"ocr_provider,omitempty"`
	AutoApproved bool    `json:"auto_approved"`
}

type KYCStatusResponse struct {
	KYCID       uint                  `json:"kyc_id"`
	Status      string                `json:"status"` // pending | approved | rejected
	SubmittedAt string                `json:"submitted_at"`
	ReviewedAt  *string               `json:"reviewed_at,omitempty"`
	ReviewNote  *string               `json:"review_note,omitempty"`
	Documents   []KYCDocumentResponse `json:"documents,omitempty"`
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
	Note string `json:"note" validate:"required"`
}

type RejectKYCRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type KYCSubmitFiles struct {
	IDFront FileBytes   // required
	Selfie  *FileBytes  // optional (หรือบังคับ ถ้าคุณเลือก)
	Others  []TypedFile // optional
}

type FileBytes struct {
	Filename string
	Bytes    []byte
}

type TypedFile struct {
	DocType  string // other
	Filename string
	Bytes    []byte
}
