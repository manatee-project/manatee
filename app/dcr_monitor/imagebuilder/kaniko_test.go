package imagebuilder

import (
	"bufio"
	"strings"
	"testing"
)

func TestGetImageAndDigestFromLog(t *testing.T) {
	// Sample log line with a typical URL and digest
	logLine := "INFO[0242] Pushed example-registry.com/namespace/repository/image@sha256:1253099ce7721d3879373d411fc7938aef80000154c9c0455c2229497ed59336\n"
	expectedImage := "example-registry.com/namespace/repository/image@sha256:1253099ce7721d3879373d411fc7938aef80000154c9c0455c2229497ed59336"
	expectedDigest := "1253099ce7721d3879373d411fc7938aef80000154c9c0455c2229497ed59336"

	// Simulate a reader with the log line
	reader := bufio.NewReader(strings.NewReader(logLine))

	// Instantiate the struct containing the function if needed
	b := KanikoImageBuilder{}

	// Call the function to test
	image, digest, err := b.getImageAndDigestFromLog(reader)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	// Verify the output matches the expected values
	if image != expectedImage {
		t.Errorf("Expected image %v, but got %v", expectedImage, image)
	}
	if digest != expectedDigest {
		t.Errorf("Expected digest %v, but got %v", expectedDigest, digest)
	}
}
