package storage

import (
	"context"
	"fmt"
	"io"
	"os"
)

type Storage interface {
	BucketPath() string
	UploadFile(reader io.Reader, remotePath string, compress bool) error
	GetFileSize(remotePath string) (int64, error)
	GetFilebyChunk(remotePath string, offset int64, chunkSize int64) ([]byte, error)
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
		storage = NewGoogleCloudStorage(ctx, getBucket())
	} else if storageType == "MINIO" {
		storage, err = NewMinioStorage(ctx, getBucket())
	}
	return storage, err
}
