package storage

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
)

type GoogleCloudStorage struct {
	ctx    context.Context
	bucket string
	client *storage.Client
}

func NewGoogleCloudStorage(ctx context.Context, bucket string) (*GoogleCloudStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage client")
	}

	return &GoogleCloudStorage{
		ctx:    ctx,
		bucket: bucket,
		client: client,
	}, nil
}

func (g *GoogleCloudStorage) Close() {
	g.client.Close()
}

func (g *GoogleCloudStorage) BucketPath() string {
	return fmt.Sprintf("gs://%s", g.bucket)
}

func (g *GoogleCloudStorage) UploadFile(reader io.Reader, remotePath string, compress bool) error {
	writer := g.client.Bucket(g.bucket).Object(remotePath).NewWriter(g.ctx)
	defer writer.Close()
	if compress {
		gzipWriter := gzip.NewWriter(writer)
		if _, err := io.Copy(gzipWriter, reader); err != nil {
			return errors.Wrap(err, "failed to copy content to gzip writer")
		}
		defer gzipWriter.Close()
	} else {
		if _, err := io.Copy(writer, reader); err != nil {
			return errors.Wrap(err, "failed to copy content to writer")
		}
	}
	return nil
}

func (g *GoogleCloudStorage) PresignedUrl(remotePath string, method string, expires time.Duration) (string, error) {
	if method != "GET" && method != "PUT" {
		return "", errors.Wrap(fmt.Errorf("unkown method for signed url, supported are GET and PUT"), "")
	}
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  method,
		Expires: time.Now().Add(expires),
	}

	return storage.SignedURL(g.bucket, remotePath, opts)
}
