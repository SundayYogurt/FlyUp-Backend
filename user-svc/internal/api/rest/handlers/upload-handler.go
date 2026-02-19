package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gofiber/fiber/v2"
)

type UploadResponse struct {
	URL      string `json:"url"`
	PublicID string `json:"public_id"`
}

type UploadHandler struct {
	cld *cloudinary.Cloudinary
}

func NewUploadHandler(cld *cloudinary.Cloudinary) *UploadHandler {
	return &UploadHandler{cld: cld}
}

func boolPtr(b bool) *bool {
	return &b
}

// POST /api/uploads/student-card
// form-data: file=<image>
func (h *UploadHandler) UploadStudentCard(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "file is required"})
	}

	// validate extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		return c.Status(400).JSON(fiber.Map{"message": "only jpg/jpeg/png/webp allowed"})
	}

	// validate size
	const maxSize = 5 * 1024 * 1024 //5MB
	if file.Size > maxSize {
		return c.Status(400).JSON(fiber.Map{"message": "file too large (max 5MB)"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "cannot open uploaded file"})
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	up, err := h.cld.Upload.Upload(ctx, f, uploader.UploadParams{
		Folder:         "flyup/student_cards",
		ResourceType:   "image",
		UseFilename:    boolPtr(true),
		UniqueFilename: boolPtr(true),
		Overwrite:      boolPtr(false),
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": fmt.Sprintf("cloudinary upload failed: %v", err)})
	}

	return c.JSON(UploadResponse{
		URL:      up.SecureURL,
		PublicID: up.PublicID,
	})
}
