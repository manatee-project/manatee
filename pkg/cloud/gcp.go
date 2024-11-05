package cloud

import (
	"compress/gzip"
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"

	"github.com/manatee-project/manatee/pkg/config"
)

type GcpService struct {
	ctx context.Context
}

// NewGcpService create gcp service
func NewGcpService(ctx context.Context) *GcpService {
	return &GcpService{ctx: ctx}
}

func (g *GcpService) DownloadFile(remoteSrcPath string, localDestPath string) error {
	client, err := storage.NewClient(g.ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create gcp storage client")
	}
	defer client.Close()

	bucket := config.GetBucket()
	objectReader, err := client.Bucket(bucket).Object(remoteSrcPath).NewReader(g.ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create reader on %s", remoteSrcPath))
	}
	defer objectReader.Close()
	f, err := os.Create(localDestPath)
	if err != nil {
		return errors.Wrap(err, "failed to create local file handler")
	}
	defer f.Close()
	if _, err = io.Copy(f, objectReader); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to download from %s to %s", remoteSrcPath, localDestPath))
	}
	return nil
}

func (g *GcpService) ListFiles(remoteDir string) ([]string, error) {
	client, err := storage.NewClient(g.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gcp storage client")
	}
	defer client.Close()
	bucket := config.GetBucket()
	it := client.Bucket(bucket).Objects(g.ctx, &storage.Query{Prefix: remoteDir})
	res := make([]string, 0)
	for {
		attrs, err := it.Next()
		if stderrors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			continue
		}
		res = append(res, attrs.Name)
	}
	return res, nil
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

func (g *GcpService) DeleteFile(remotePath string) error {
	client, err := storage.NewClient(g.ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create storage client")
	}
	defer client.Close()
	bucket := config.GetBucket()
	if err := client.Bucket(bucket).Object(remotePath).Delete(g.ctx); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to delete cloud storage object: %s/%s", bucket, remotePath))
	}
	return nil
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

func (g *GcpService) GetServiceAccountEmail() (string, error) {
	if metadata.OnGCE() {
		email, err := metadata.Email("default")
		if err != nil {
			return "", errors.Wrap(err, "failed to get service account email")
		}
		return email, nil
	}
	return "", nil
}
