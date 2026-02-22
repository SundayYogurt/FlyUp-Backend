package iapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type FaceMatchResult struct {
	Confidence   float64 `json:"confidence"`
	IsSamePerson string  `json:"isSamePerson"` // docs เป็น string "true"/"false"
}

type FaceAndIDCardVerificationResponse struct {
	IDCard      FaceMatchResult `json:"idcard"`
	Selfie      FaceMatchResult `json:"selfie"`
	Total       FaceMatchResult `json:"total"`
	TimeProcess float64         `json:"time_process"`
	// เผื่อกรณี error เป็น json
	ErrorMessage string `json:"error_message,omitempty"`
}

type Client struct {
	apiKey string
	http   *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// VerifyFaceAndIDCard: เทียบ "รูปบัตรประชาชน"กับ "รูปเซลฟี่"
// Endpoint: POST /v3/store/ekyc/face-and-id-card-verification
func (c *Client) VerifyFaceAndIDCard(
	ctx context.Context,
	idCardFilename string,
	idCardReader io.Reader,
	selfieFilename string,
	selfieReader io.Reader,
) (*FaceAndIDCardVerificationResponse, error) {
	if c.apiKey == "" {
		return nil, errors.New("missing iapp api key")
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// file0 = Selfie image
	fw0, err := w.CreateFormFile("file1", idCardFilename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw0, idCardReader); err != nil {
		return nil, err
	}

	// file1 =ID card image
	fw1, err := w.CreateFormFile("file0", selfieFilename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw1, selfieReader); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	url := "https://api.iapp.co.th/v3/store/ekyc/face-and-id-card-verification"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("apikey", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// non-2xx -> คืน error ออกไปเลย (พยายามดึง error_message ถ้าเป็น json)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e FaceAndIDCardVerificationResponse
		if json.Unmarshal(body, &e) == nil && e.ErrorMessage != "" {
			return &e, fmt.Errorf("iapp error (%d): %s", resp.StatusCode, e.ErrorMessage)
		}
		return nil, fmt.Errorf("iapp http error (%d): %s", resp.StatusCode, string(body))
	}

	var out FaceAndIDCardVerificationResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	// เผื่อบางเคส iApp ส่ง error_message มากับ 200
	if out.ErrorMessage != "" {
		return &out, errors.New(out.ErrorMessage)
	}

	return &out, nil
}
