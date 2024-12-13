package storage

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	credentialspb "cloud.google.com/go/iam/credentials/apiv1/credentialspb"
)

type GoogleCloudStorage struct {
	ctx            context.Context
	bucket         string
	client         *storage.Client
	googleAccessId string
}

func NewGoogleCloudStorage(ctx context.Context, bucket string) (*GoogleCloudStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage client")
	}
	serviceAccount, err := getGoogleServiceAccount()

	return &GoogleCloudStorage{
		ctx:            ctx,
		bucket:         bucket,
		client:         client,
		googleAccessId: serviceAccount,
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
	c, err := credentials.NewIamCredentialsClient(g.ctx)
	if err != nil {
		panic(err)
	}

	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         method,
		Expires:        time.Now().Add(expires),
		GoogleAccessID: g.googleAccessId,
		SignBytes: func(b []byte) ([]byte, error) {
			req := &credentialspb.SignBlobRequest{
				Payload: b,
				Name:    g.googleAccessId,
			}
			resp, err := c.SignBlob(g.ctx, req)
			if err != nil {
				return nil, errors.Wrap(err, "failed to sign blocb")
			}
			return resp.SignedBlob, err
		},
	}
	url, err := storage.SignedURL(g.bucket, remotePath, opts)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign url")
	}
	return url, nil
}

func getGoogleServiceAccount() (string, error) {
	url := "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/email"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create http client")
	}
	req.Header.Add("Metadata-Flavor", "Google")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to request google meta service account")
	}
	defer resp.Body.Close()
	// 读取响应体
	account, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to request google meta service account")
	}
	return string(account), nil
}
