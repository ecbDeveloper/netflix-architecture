package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Service interface {
	Upload(ctx context.Context, fileKey string, file io.Reader) error
	DeleteFile(ctx context.Context, fileKey string) error
}

type ServiceImpl struct {
	S3Client    *s3.Client
	BucketName  string
	EndpointURL string
}

func NewService(s3Client *s3.Client, bucketName string, endpointURL string) Service {
	return &ServiceImpl{
		S3Client:    s3Client,
		BucketName:  bucketName,
		EndpointURL: endpointURL,
	}
}

func (s *ServiceImpl) Upload(ctx context.Context, fileKey string, file io.Reader) error {
	_, err := s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(fileKey),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}

func (s *ServiceImpl) DeleteFile(ctx context.Context, fileKey string) error {
	key := s.extractKey(fileKey)

	_, err := s.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

func (s *ServiceImpl) extractKey(objectURL string) string {
	parsed, err := url.Parse(objectURL)
	if err != nil || parsed.Scheme == "" {
		return objectURL
	}

	prefix := "/" + s.BucketName + "/"
	if strings.HasPrefix(parsed.Path, prefix) {
		return strings.TrimPrefix(parsed.Path, prefix)
	}

	return strings.TrimPrefix(parsed.Path, "/")
}
