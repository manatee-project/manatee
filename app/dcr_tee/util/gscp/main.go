package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	storageBucket "github.com/manatee-project/manatee/pkg/storage"
)

// a util to replace `gsutil cp` command, to remove unncessary installation of a heavy debian package
func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <local-file-path> <remote-path (gs://bucket-name/path/to/dir/)>\n", os.Args[0])
	}

	localFilePath := os.Args[1]
	gcsURL := os.Args[2]

	ctx := context.Background()
	storage, object, err := extractBucketAndObject(ctx, gcsURL)
	if err != nil {
		log.Fatalf("invalid URL: %v", err)
	}

	file, err := os.Open(localFilePath)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	err = storage.UploadFile(file, object, false)
	if err != nil {
		log.Fatalf("failed to upload file: %v", err)
	}

	log.Printf("File %s uploaded to gs://%s/%s successfully.\n", localFilePath, storage.BucketPath(), object)
}

func extractBucketAndObject(ctx context.Context, gcsURL string) (storage storageBucket.Storage, objectPath string, err error) {
	u, err := url.Parse(gcsURL)
	if err != nil {
		return nil, "", fmt.Errorf("invalid GCS URL: %v", err)
	}

	if u.Scheme != "gs" && u.Scheme != "s3" {
		return nil, "", fmt.Errorf("URL must start with gs:// or s3://")
	}

	bucketName := u.Host
	objectPath = strings.TrimPrefix(u.Path, "/")

	if u.Scheme == "gs" {
		storage = storageBucket.NewGoogleCloudStorage(ctx, bucketName)
	} else if u.Scheme == "s3" {
		storage, err = storageBucket.NewMinioStorage(ctx, bucketName)
		if err != nil {
			return nil, objectPath, err
		}
	}

	return storage, objectPath, nil
}
