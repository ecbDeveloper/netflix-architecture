package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/config"
)

type S3Client struct {
	client     *s3.Client
	bucketName string
}

func NewS3Client(ctx context.Context, cfg *config.Config) (*S3Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	return &S3Client{
		client: s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.S3EndpointURL)
			o.UsePathStyle = true
		}),
		bucketName: cfg.S3BucketName,
	}, nil
}

func (s *S3Client) DownloadFile(ctx context.Context, key string, localPath string) error {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download from s3: %w", err)
	}
	defer out.Body.Close()

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, out.Body)
	if err != nil {
		return fmt.Errorf("failed to write to local file: %w", err)
	}

	return nil
}

func (s *S3Client) UploadFile(ctx context.Context, localPath string, key string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to s3: %w", err)
	}

	return nil
}

func (s *S3Client) UploadDirectory(ctx context.Context, localDir string, s3Prefix string) error {
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return fmt.Errorf("failed to read local directory: %w", err)
	}

	for _, entry := range entries {
		localPath := filepath.Join(localDir, entry.Name())
		s3Key := fmt.Sprintf("%s/%s", s3Prefix, entry.Name())

		if entry.IsDir() {
			if err := s.UploadDirectory(ctx, localPath, s3Key); err != nil {
				return err
			}
		} else {
			if err := s.UploadFile(ctx, localPath, s3Key); err != nil {
				return err
			}
		}
	}
	return nil
}
