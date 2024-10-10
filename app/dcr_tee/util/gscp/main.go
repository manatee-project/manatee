package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
)

// a util to replace `gsutil cp` command, to remove unncessary installation of a heavy debian package
func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <local-file-path> <remote-path (gs://bucket-name/path/to/dir/)>\n", os.Args[0])
	}

	localFilePath := os.Args[1]
	gcsURL := os.Args[2]

	bucket, object, err := extractBucketAndObject(gcsURL)
	if err != nil {
		log.Fatalf("invalid GCS URL: %v", err)
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	file, err := os.Open(localFilePath)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		log.Fatalf("failed to copy file to GCS: %v", err)
	}

	if err := wc.Close(); err != nil {
		log.Fatalf("failed to close GCS object: %v", err)
	}

	log.Printf("File %s uploaded to gs://%s/%s successfully.\n", localFilePath, bucket, object)
}

func extractBucketAndObject(gcsURL string) (bucketName, objectPath string, err error) {
	u, err := url.Parse(gcsURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid GCS URL: %v", err)
	}

	if u.Scheme != "gs" {
		return "", "", fmt.Errorf("URL must start with gs://")
	}

	bucketName = u.Host
	objectPath = strings.TrimPrefix(u.Path, "/")

	// Extract the filename from the local file path

	return bucketName, objectPath, nil
}
