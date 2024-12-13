package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
)

type MinioStorage struct {
	ctx         context.Context
	bucket      string
	minioClient minio.Client
}

func NewMinioStorage(ctx context.Context, bucket string) (*MinioStorage, error) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKeyID == "" {
		panic("AWS_ACCESS_KEY_ID environment variable is not present")
	}
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		panic("AWS_SECRET_ACCESS_KEY environment variable is not present")
	}
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		panic("S3_ENDPOINT environment variable is not present")
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	exist, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}

	if !exist {
		err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us"})
		if err != nil {
			return nil, err
		}
	}

	return &MinioStorage{
		ctx:         ctx,
		bucket:      bucket,
		minioClient: *minioClient,
	}, nil
}

func (m *MinioStorage) Close() {
}

func (m *MinioStorage) BucketPath() string {
	return fmt.Sprintf("s3://%s", m.bucket)
}

// compress parameter hasn't been implemented for minio client
func (m *MinioStorage) UploadFile(reader io.Reader, remotePath string, compress bool) error {
	_, err := m.minioClient.PutObject(m.ctx, m.bucket, remotePath, reader, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return errors.Wrap(err, "failed to upload to minio")
	}
	return nil
}

func (m *MinioStorage) PresignedUrl(remotePath string, method string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	var url *url.URL
	var err error
	if method == "GET" {
		url, err = m.minioClient.PresignedGetObject(m.ctx, m.bucket, remotePath, expires, reqParams)
		if err != nil {
			return "", err
		}
	} else if method == "PUT" {
		url, err = m.minioClient.PresignedPutObject(m.ctx, m.bucket, remotePath, expires)
		if err != nil {
			return "", err
		}
	} else {
		return "", errors.Wrap(fmt.Errorf("unkown method for signed url, supported are GET and PUT"), "")
	}
	return url.String(), nil
}
