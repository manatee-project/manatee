package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
)

type MinioStorage struct {
	ctx             context.Context
	bucket          string
	endpoint        string
	accessKeyID     string
	secretAccessKey string
}

func NewMinioStorage(ctx context.Context, bucket string) *MinioStorage {
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
	return &MinioStorage{
		ctx:             ctx,
		bucket:          bucket,
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
	}
}

func (m *MinioStorage) BucketPath() string {
	return fmt.Sprintf("s3://%s", m.bucket)
}

// compress parameter hasn't been implemented for minio client
func (m *MinioStorage) UploadFile(reader io.Reader, remotePath string, compress bool) error {
	minioClient, err := minio.New(m.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.accessKeyID, m.secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create minio client")
	}
	_, err = minioClient.PutObject(m.ctx, m.bucket, remotePath, reader, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return errors.Wrap(err, "failed to upload to minio")
	}
	return nil
}

func (m *MinioStorage) GetFileSize(remotePath string) (int64, error) {
	minioClient, err := minio.New(m.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.accessKeyID, m.secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return 0, errors.Wrap(err, "failed to create minio client")
	}
	objectInfo, err := minioClient.StatObject(m.ctx, m.bucket, remotePath, minio.StatObjectOptions{})
	if err != nil {
		return 0, errors.Wrap(err, "failed to stat minio object")
	}
	return objectInfo.Size, nil
}

func (m *MinioStorage) GetFilebyChunk(remotePath string, offset int64, chunkSize int64) ([]byte, error) {
	minioClient, err := minio.New(m.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.accessKeyID, m.secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create minio client")
	}
	object, err := minioClient.GetObject(context.Background(), m.bucket, remotePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get minio oibject")
	}
	buf := make([]byte, chunkSize)
	_, err = object.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "failed to get minio oibject")
	}
	return buf, nil
}
