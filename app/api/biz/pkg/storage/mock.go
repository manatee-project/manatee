package storage

import (
	"context"
	"io"
	"time"
)

// A mock storage only used for job service testing
type MockStorage struct {
	ctx context.Context
}

func NewMockStorage(ctx context.Context) *MockStorage {
	return &MockStorage{
		ctx: ctx,
	}
}

func (m *MockStorage) Close() {
}

func (m *MockStorage) BucketPath() string {
	return ""
}

func (m *MockStorage) UploadFile(reader io.Reader, remotePath string, compress bool) error {
	return nil
}

func (m *MockStorage) IssueSignedUrl(remotePath string, method string, expires time.Duration) (string, error) {
	return "", nil
}
