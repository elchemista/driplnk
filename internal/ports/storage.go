package ports

import (
	"context"
	"io"
)

type FileUploader interface {
	// Upload uploads a file as-is.
	Upload(ctx context.Context, file io.Reader, filename string) (string, error)
	// UploadProfileImage processes (resizes/converts) and uploads a profile image.
	UploadProfileImage(ctx context.Context, file io.Reader, filename string) (string, error)
}
