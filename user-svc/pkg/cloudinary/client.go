package cloudinary

import (
	"github.com/cloudinary/cloudinary-go/v2"
)

func New() (*cloudinary.Cloudinary, error) {
	// cloudinary.New() จะอ่านจาก CLOUDINARY_URL ถ้ามี
	// แต่คุณตั้งเป็น 3 ตัวแยก ก็ใช้ NewFromParams จะชัวร์กว่า
	return cloudinary.New()
}
