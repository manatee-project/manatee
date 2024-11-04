package cloud

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"

	"github.com/manatee-project/manatee/pkg/config"
)

type WorkloadIdentityPoolProvider struct {
	DisplayName        string            `json:"displayName"`
	Description        string            `json:"description"`
	AttributeMapping   map[string]string `json:"attributeMapping"`
	AttributeCondition string            `json:"attributeCondition"`
	OIDC               OIDC              `json:"oidc"`
}

type OIDC struct {
	IssuerUri        string   `json:"issuerUri"`
	AllowedAudiences []string `json:"allowedAudiences"`
	JwksJson         string   `json:"jwksJson"`
}

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   uint64 `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

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

func workloadIdentityRequestBody(name string, imageDigest string) ([]byte, error) {
	serviceAccountEmail := config.GetCvmServiceAccountEmail()
	attributeCondition := fmt.Sprintf("assertion.submods.container.image_digest == '%s' && '%s' in assertion.google_service_accounts && assertion.swname == 'CONFIDENTIAL_SPACE'", imageDigest, serviceAccountEmail)
	if !config.IsDebug() {
		attributeCondition += " && 'STABLE' in assertion.submods.confidential_space.support_attributes"
	}
	provider := WorkloadIdentityPoolProvider{
		DisplayName: name,
		OIDC: OIDC{
			IssuerUri:        "https://confidentialcomputing.googleapis.com/",
			AllowedAudiences: []string{"https://sts.googleapis.com"},
		},
		AttributeMapping: map[string]string{
			"google.subject": "assertion.sub",
		},
		AttributeCondition: attributeCondition,
	}
	jsonStringBytes, err := json.Marshal(provider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal workload identity pool provider")
	}
	return jsonStringBytes, nil
}

func (g *GcpService) getAccessToken() (string, error) {
	// Get access token
	if metadata.OnGCE() {
		res, err := metadata.Get("instance/service-accounts/default/token")
		if err != nil {
			return "", err
		}
		var token Token
		if err = json.Unmarshal([]byte(res), &token); err != nil {
			return "", err
		}
		return token.AccessToken, nil
	}
	// Run on minikube
	// Use application credentials to get token
	credentials, err := google.FindDefaultCredentials(g.ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", errors.Wrap(err, "failed to find default credential")
	}
	token, err := credentials.TokenSource.Token()
	if err != nil {
		return "", errors.Wrap(err, "failed to get token source")
	}
	return token.AccessToken, nil
}

func (g *GcpService) CreateWorkloadIdentityPoolProvider(name string) error {
	requestBody, err := workloadIdentityRequestBody(name, "")
	if err != nil {
		// err already is wrapped
		return err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", config.GetCreateWipProviderUrl(name), bytes.NewReader(requestBody))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	token, err := g.getAccessToken()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to do http request")
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read http response")
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

func (g *GcpService) PrepareResourcesForUser(user string) error {
	// create the workload identity pool provider for the user. The image digest set empty
	wipProvider := config.GetUserWipProvider(user)
	err := g.CreateWorkloadIdentityPoolProvider(wipProvider)
	if err != nil {
		return err
	}

	return nil
}
