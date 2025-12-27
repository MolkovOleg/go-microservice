package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-microservice/models"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type IntegrationService struct {
	client        *minio.Client
	defaultBucket string
}

func NewIntegrationService(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*IntegrationService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	svc := &IntegrationService{
		client:        client,
		defaultBucket: bucket,
	}
	if bucket != "" {
		if err := svc.ensureBucket(context.Background(), bucket); err != nil {
			log.Printf("WARN: unable to ensure bucket %s: %v", bucket, err)
		}
	}
	return svc, nil
}
func (s *IntegrationService) ensureBucket(ctx context.Context, bucket string) error {
	if bucket == "" {
		return fmt.Errorf("bucket name is required")
	}
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}
func (s *IntegrationService) bucketOrDefault(bucket string) (string, error) {
	if bucket != "" {
		return bucket, nil
	}
	if s.defaultBucket != "" {
		return s.defaultBucket, nil
	}
	return "", fmt.Errorf("bucket name is required")
}
func (s *IntegrationService) UploadObject(ctx context.Context, bucket, objectName string, content []byte, contentType string) (*minio.UploadInfo, error) {
	targetBucket, err := s.bucketOrDefault(bucket)
	if err != nil {
		return nil, err
	}
	if objectName == "" {
		return nil, fmt.Errorf("object name is required")
	}
	if err := s.ensureBucket(ctx, targetBucket); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(content)
	info, err := s.client.PutObject(ctx, targetBucket, objectName, reader, reader.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, err
	}
	return &info, nil
}
func (s *IntegrationService) PresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	targetBucket, err := s.bucketOrDefault(bucket)
	if err != nil {
		return "", err
	}
	if objectName == "" {
		return "", fmt.Errorf("object name is required")
	}
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}
	if err := s.ensureBucket(ctx, targetBucket); err != nil {
		return "", err
	}
	url, err := s.client.PresignedGetObject(ctx, targetBucket, objectName, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
func (s *IntegrationService) ExportUserSnapshot(ctx context.Context, bucket string, user *models.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user is required")
	}
	payload, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	objectName := fmt.Sprintf("users/user-%d-%d.json", user.ID, time.Now().UnixNano())
	if _, err := s.UploadObject(ctx, bucket, objectName, payload, "application/json"); err != nil {
		return "", err
	}
	return objectName, nil
}