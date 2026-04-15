package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
)

type Service interface {
	Upload(ctx context.Context, fileName string, file io.Reader) (string, error)
}

type ServiceImpl struct {
	uploadPath string
}

func NewService(uploadPath string) Service {
	return &ServiceImpl{
		uploadPath: uploadPath,
	}
}

func (s *ServiceImpl) Upload(ctx context.Context, fileName string, file io.Reader) (string, error) {
	contentURL := path.Join(s.uploadPath, fileName)

	dst, err := os.Create(contentURL)
	if err != nil {
		return "", fmt.Errorf("failed to create file path: %w", err)
	}
	defer dst.Close()

	io.Copy(dst, file)

	return contentURL, nil
}
