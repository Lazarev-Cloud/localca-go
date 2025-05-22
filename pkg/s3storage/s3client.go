package s3storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Client wraps the MinIO client for S3 operations
type S3Client struct {
	client     *minio.Client
	bucketName string
	enabled    bool
}

// NewS3Client creates a new S3 client
func NewS3Client(cfg *config.Config) (*S3Client, error) {
	if !cfg.S3Enabled {
		return &S3Client{enabled: false}, nil
	}

	// Initialize MinIO client
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
		Region: cfg.S3Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	s3Client := &S3Client{
		client:     client,
		bucketName: cfg.S3BucketName,
		enabled:    true,
	}

	// Ensure bucket exists
	if err := s3Client.ensureBucket(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return s3Client, nil
}

// IsEnabled returns whether S3 storage is enabled
func (s *S3Client) IsEnabled() bool {
	return s.enabled
}

// ensureBucket creates the bucket if it doesn't exist
func (s *S3Client) ensureBucket() error {
	if !s.enabled {
		return nil
	}

	ctx := context.Background()

	// Check if bucket exists
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		// Create bucket
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile uploads a file to S3
func (s *S3Client) UploadFile(objectName string, data []byte, contentType string) error {
	if !s.enabled {
		return fmt.Errorf("S3 storage is not enabled")
	}

	ctx := context.Background()
	reader := bytes.NewReader(data)

	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %s: %w", objectName, err)
	}

	return nil
}

// DownloadFile downloads a file from S3
func (s *S3Client) DownloadFile(objectName string) ([]byte, error) {
	if !s.enabled {
		return nil, fmt.Errorf("S3 storage is not enabled")
	}

	ctx := context.Background()

	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", objectName, err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s: %w", objectName, err)
	}

	return data, nil
}

// DeleteFile deletes a file from S3
func (s *S3Client) DeleteFile(objectName string) error {
	if !s.enabled {
		return fmt.Errorf("S3 storage is not enabled")
	}

	ctx := context.Background()

	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", objectName, err)
	}

	return nil
}

// ListFiles lists files in S3 with a given prefix
func (s *S3Client) ListFiles(prefix string) ([]string, error) {
	if !s.enabled {
		return nil, fmt.Errorf("S3 storage is not enabled")
	}

	ctx := context.Background()
	var files []string

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}

// FileExists checks if a file exists in S3
func (s *S3Client) FileExists(objectName string) (bool, error) {
	if !s.enabled {
		return false, fmt.Errorf("S3 storage is not enabled")
	}

	ctx := context.Background()

	_, err := s.client.StatObject(ctx, s.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		// Check if error is "object not found"
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat object %s: %w", objectName, err)
	}

	return true, nil
}

// GetFileInfo returns file information
func (s *S3Client) GetFileInfo(objectName string) (*minio.ObjectInfo, error) {
	if !s.enabled {
		return nil, fmt.Errorf("S3 storage is not enabled")
	}

	ctx := context.Background()

	info, err := s.client.StatObject(ctx, s.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", objectName, err)
	}

	return &info, nil
}

// GeneratePresignedURL generates a presigned URL for file access
func (s *S3Client) GeneratePresignedURL(objectName string, expiry time.Duration) (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("S3 storage is not enabled")
	}

	ctx := context.Background()

	url, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL for %s: %w", objectName, err)
	}

	return url.String(), nil
}

// UploadCertificateFiles uploads all certificate-related files
func (s *S3Client) UploadCertificateFiles(certName string, files map[string][]byte) error {
	if !s.enabled {
		return fmt.Errorf("S3 storage is not enabled")
	}

	for filename, data := range files {
		objectName := s.getCertificateObjectName(certName, filename)
		contentType := s.getContentType(filename)

		if err := s.UploadFile(objectName, data, contentType); err != nil {
			return fmt.Errorf("failed to upload %s: %w", filename, err)
		}
	}

	return nil
}

// DownloadCertificateFiles downloads all certificate-related files
func (s *S3Client) DownloadCertificateFiles(certName string) (map[string][]byte, error) {
	if !s.enabled {
		return nil, fmt.Errorf("S3 storage is not enabled")
	}

	prefix := fmt.Sprintf("certificates/%s/", certName)
	files, err := s.ListFiles(prefix)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for _, objectName := range files {
		data, err := s.DownloadFile(objectName)
		if err != nil {
			return nil, err
		}

		// Extract filename from object name
		filename := filepath.Base(objectName)
		result[filename] = data
	}

	return result, nil
}

// DeleteCertificateFiles deletes all certificate-related files
func (s *S3Client) DeleteCertificateFiles(certName string) error {
	if !s.enabled {
		return fmt.Errorf("S3 storage is not enabled")
	}

	prefix := fmt.Sprintf("certificates/%s/", certName)
	files, err := s.ListFiles(prefix)
	if err != nil {
		return err
	}

	for _, objectName := range files {
		if err := s.DeleteFile(objectName); err != nil {
			return err
		}
	}

	return nil
}

// getCertificateObjectName generates the S3 object name for a certificate file
func (s *S3Client) getCertificateObjectName(certName, filename string) string {
	return fmt.Sprintf("certificates/%s/%s", certName, filename)
}

// getContentType returns the appropriate content type for a file
func (s *S3Client) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".crt", ".pem":
		return "application/x-pem-file"
	case ".key":
		return "application/x-pem-file"
	case ".p12", ".pfx":
		return "application/x-pkcs12"
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
