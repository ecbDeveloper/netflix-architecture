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
	Upload(ctx context.Context, fileName string, file io.Reader) (string, error)
	DeleteFile(ctx context.Context, objectURL string) error
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

// Upload envia o arquivo para o bucket S3 e retorna a URL pública do objeto.
func (s *ServiceImpl) Upload(ctx context.Context, fileName string, file io.Reader) (string, error) {
	_, err := s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	objectURL := fmt.Sprintf("%s/%s/%s", s.EndpointURL, s.BucketName, fileName)

	return objectURL, nil
}

// DeleteFile remove o objeto do bucket S3. Aceita tanto a URL pública completa
// quanto apenas a chave (nome do arquivo) do objeto.
func (s *ServiceImpl) DeleteFile(ctx context.Context, objectURL string) error {
	key := s.extractKey(objectURL)

	_, err := s.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

// extractKey extrai a chave do objeto a partir de uma URL pública completa ou
// retorna o valor original caso já seja uma chave simples.
func (s *ServiceImpl) extractKey(objectURL string) string {
	parsed, err := url.Parse(objectURL)
	if err != nil || parsed.Scheme == "" {
		// Não é uma URL válida — assume que já é a chave.
		return objectURL
	}

	// Caminho esperado: /<bucket>/<key>  (path-style)
	prefix := "/" + s.BucketName + "/"
	if strings.HasPrefix(parsed.Path, prefix) {
		return strings.TrimPrefix(parsed.Path, prefix)
	}

	// Fallback: usa o path inteiro sem a barra inicial.
	return strings.TrimPrefix(parsed.Path, "/")
}
