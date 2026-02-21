package iapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type ThaiIDFrontResponse struct {
	IDNumber       string  `json:"id_number"`
	DetectionScore float64 `json:"detection_score"`
	ErrorMessage   string  `json:"error_message"`
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

func (c *Client) OCRThaiIDFront(ctx context.Context, filename string, r io.Reader) (*ThaiIDFrontResponse, error) {
	if c.apiKey == "" {
		return nil, errors.New("missing iapp api key")
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw, r); err != nil {
		return nil, err
	}

	// ปิด writer เพื่อเขียน boundary ให้ครบ
	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.iapp.co.th/v3/store/ekyc/thai-national-id-card/front",
		&buf,
	)
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(string(body))
	}

	var out ThaiIDFrontResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	// ถ้า iApp ส่ง error_message มา
	if out.ErrorMessage != "" {
		return &out, errors.New(out.ErrorMessage)
	}

	return &out, nil
}
