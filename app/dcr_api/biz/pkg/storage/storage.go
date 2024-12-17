package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
)

type Storage interface {
	BucketPath() string
	UploadFile(reader io.Reader, remotePath string, compress bool) error
	IssueSignedUrl(remotePath string, method string, expiry time.Duration) (string, error)
	Close()
}

func getBucket() (string, error) {
	env := os.Getenv("ENV")
	if env == "" {
		return "", errors.Wrap(fmt.Errorf("ENV environment variable is not present"), "")
	}
	return fmt.Sprintf("dcr-%s-hub", env), nil
}

func GetStorage(ctx context.Context) (Storage, error) {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "GCP"
	}
	var storage Storage
	bucket, err := getBucket()
	if err != nil {
		return storage, err
	}
	if storageType == "GCP" {
		storage, err = NewGoogleCloudStorage(ctx, bucket)
	} else if storageType == "MINIO" {
		storage, err = NewMinioStorage(ctx, bucket)
	}
	return storage, err
}
