package cloudinary

import (
	"github.com/cloudinary/cloudinary-go/v2"
)

func New() (*cloudinary.Cloudinary, error) {
	return cloudinary.New()
}
