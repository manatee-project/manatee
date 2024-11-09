package cloud

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"

	"github.com/manatee-project/manatee/pkg/config"
)

type GcpService struct {
	ctx context.Context
}

// NewGcpService create gcp service
func NewGcpService(ctx context.Context) *GcpService {
	return &GcpService{ctx: ctx}
}

func (g *GcpService) GetFileSize(remotePath string) (int64, error) {
	client, err := storage.NewClient(g.ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create gcp storage client")
	}
	defer client.Close()
	bucket := config.GetBucket()
	attr, err := client.Bucket(bucket).Object(remotePath).Attrs(g.ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get file attributes, or it doesn't exist")
	}
	return attr.Size, nil
}

func (g *GcpService) GetFilebyChunk(remotePath string, offset int64, chunkSize int64) ([]byte, error) {
	client, err := storage.NewClient(g.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gcp storage client")
	}
	defer client.Close()
	bucket := config.GetBucket()
	objectHandle := client.Bucket(bucket).Object(remotePath)
	objectReader, err := objectHandle.NewRangeReader(g.ctx, offset, chunkSize)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to create reader on %s", remotePath))
	}
	defer objectReader.Close()
	data := make([]byte, chunkSize)
	n, err := objectReader.Read(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cloud storage object")
	}
	data = data[:n]
	return data, nil
}

func (g *GcpService) UploadFile(reader io.Reader, remotePath string, compress bool) error {
	client, err := storage.NewClient(g.ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create storage client")
	}
	defer client.Close()
	bucket := config.GetBucket()
	writer := client.Bucket(bucket).Object(remotePath).NewWriter(g.ctx)
	defer writer.Close()
	if compress {
		gzipWriter := gzip.NewWriter(writer)
		if _, err = io.Copy(gzipWriter, reader); err != nil {
			return errors.Wrap(err, "failed to copy content to gzip writer")
		}
		defer gzipWriter.Close()
	} else {
		if _, err = io.Copy(writer, reader); err != nil {
			return errors.Wrap(err, "failed to copy content to writer")
		}
	}
	return nil
}
