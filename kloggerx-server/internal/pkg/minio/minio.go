package miniosvc

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"kloggerx-server/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var client *minio.Client
var isInitialized bool

// Init initializes MinIO client
func Init(cfg config.MinIOConfig) error {
	var err error
	client, err = minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	isInitialized = true
	return nil
}

// IsAvailable returns true if MinIO is initialized and configuration is usable
func IsAvailable() bool {
	return isInitialized && client != nil && config.Cfg.MinIO.Bucket != ""
}

// Upload uploads a file to storage (MinIO or local filesystem as fallback)
func Upload(objectName string, reader io.Reader, size int64, contentType string) error {
	if IsAvailable() {
		_, err := client.PutObject(context.Background(), config.Cfg.MinIO.Bucket, objectName, reader, size, minio.PutObjectOptions{
			ContentType: contentType,
		})
		return err
	}

	// Fallback to local filesystem
	return uploadToLocal(objectName, reader, size)
}

// uploadToLocal saves file to local filesystem
func uploadToLocal(objectName string, reader io.Reader, size int64) error {
	// Create uploads directory
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create uploads directory: %w", err)
	}

	// Create full path
	fullPath := filepath.Join(uploadDir, objectName)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, reader)
	return err
}

// GetPresignedURL returns a URL for accessing the file
func GetPresignedURL(objectName string) (string, error) {
	if IsAvailable() {
		url, err := client.PresignedGetObject(context.Background(), config.Cfg.MinIO.Bucket, objectName, 24*time.Hour, nil)
		if err == nil {
			return url.String(), nil
		}
		// Degrade gracefully when MinIO client/config is in bad state
		return fmt.Sprintf("/uploads/%s", objectName), nil
	}

	// Fallback: return local file path as URL
	return fmt.Sprintf("/uploads/%s", objectName), nil
}

// Delete removes a file from storage
func Delete(objectName string) error {
	if IsAvailable() {
		return client.RemoveObject(context.Background(), config.Cfg.MinIO.Bucket, objectName, minio.RemoveObjectOptions{})
	}

	// Fallback: delete from local filesystem
	fullPath := filepath.Join("uploads", objectName)
	return os.Remove(fullPath)
}

// GetFileContent reads file content from storage
func GetFileContent(objectName string) ([]byte, error) {
	if IsAvailable() {
		obj, err := client.GetObject(context.Background(), config.Cfg.MinIO.Bucket, objectName, minio.GetObjectOptions{})
		if err == nil {
			defer obj.Close()
			if content, readErr := io.ReadAll(obj); readErr == nil {
				return content, nil
			}
		}
		// Degrade gracefully to local filesystem when MinIO read fails
	}

	// Fallback: read from local filesystem
	fullPath := filepath.Join("uploads", objectName)
	return os.ReadFile(fullPath)
}

// GetObject returns a ReadCloser for streaming a file from storage.
func GetObject(objectName string) (io.ReadCloser, error) {
	if IsAvailable() {
		obj, err := client.GetObject(context.Background(), config.Cfg.MinIO.Bucket, objectName, minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	// Fallback: open from local filesystem
	fullPath := filepath.Join("uploads", objectName)
	f, err := os.Open(fullPath)
	return f, err
}
