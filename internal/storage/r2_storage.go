package storage

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Storage struct {
	Client    *s3.Client
	Bucket    string
	PublicURL string
}

func NewR2Storage(ctx context.Context, endpoint, accessKey, secretKey, bucket, publicURL, region string) (*R2Storage, error) {
	if region == "" {
		region = "auto"
	}
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           endpoint,
			PartitionID:   "aws",
			SigningName:   "s3",
			SigningRegion: region,
		}, nil
	})
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("gagal load config S3: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	return &R2Storage{
		Client:    client,
		Bucket:    bucket,
		PublicURL: publicURL,
	}, nil
}

func (s *R2Storage) Save(ctx context.Context, reportType, filename string, file *bytes.Buffer) (string, error) {
	objectKey := fmt.Sprintf("%s/%s", reportType, filename)

	ext := strings.ToLower(filepath.Ext(filename))
	contentType := "application/octet-stream"
	switch ext {
	case ".pdf":
		contentType = "application/pdf"
	case ".webp":
		contentType = "image/webp"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	}

	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("gagal upload ke R2: %w", err)
	}
	publicURL := fmt.Sprintf("%s/%s", s.PublicURL, objectKey)
	return publicURL, nil
}
