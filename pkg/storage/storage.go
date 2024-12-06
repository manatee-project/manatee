package storage

import "io"

type Storage interface {
	BucketPath(bucketName string) string
	UploadFile(reader io.Reader, remotePath string, compress bool) error
	GetFileSize(remotePath string) (int64, error)
	GetFilebyChunk(remotePath string, offset int64, chunkSize int64) ([]byte, error)
}
