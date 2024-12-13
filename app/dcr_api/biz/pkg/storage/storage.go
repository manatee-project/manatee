package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

type Storage interface {
	BucketPath() string
	UploadFile(reader io.Reader, remotePath string, compress bool) error
	PresignedUrl(remotePath string, method string, expiry time.Duration) (string, error)
	Close()
}

func getBucket() string {
	env := os.Getenv("ENV")
	if env == "" {
		panic("ENV environment variable is not present")
	}
	return fmt.Sprintf("dcr-%s-hub", env)
}

func GetStorage(ctx context.Context) (Storage, error) {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "GCP"
	}
	var storage Storage
	var err error
	if storageType == "GCP" {
		storage, err = NewGoogleCloudStorage(ctx, getBucket())
	} else if storageType == "MINIO" {
		storage, err = NewMinioStorage(ctx, getBucket())
	}
	return storage, err
}
