package cloudinary

import (
	"bytes"
	"context"

	cld "github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryUploader struct {
	cld *cld.Cloudinary
}

func NewCloudinaryUploader(cloud *cld.Cloudinary) *CloudinaryUploader {
	return &CloudinaryUploader{cld: cloud}
}

func (u *CloudinaryUploader) UploadBytes(
	ctx context.Context,
	folder string,
	filename string,
	b []byte,
) (string, error) {

	// ต้องส่งเป็น io.Reader ไม่ใช่ []byte ตรงๆ
	reader := bytes.NewReader(b)

	res, err := u.cld.Upload.Upload(
		ctx,
		reader,
		uploader.UploadParams{
			Folder:       folder,
			PublicID:     filename,
			ResourceType: "image", // ช่วยชัวร์ว่าเป็นรูป
		},
	)
	if err != nil {
		return "", err
	}

	return res.SecureURL, nil
}
