package interfaces

import "context"

type Uploader interface {
	UploadBytes(ctx context.Context, folder string, filename string, b []byte) (string, error)
}
