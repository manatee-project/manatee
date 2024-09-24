package cloud

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/iam/apiv1/iampb"
	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"cloud.google.com/go/storage"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

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

func (g *GcpService) CreateSymmetricKeys(keyId string) error {
	client, err := kms.NewKeyManagementClient(g.ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create key management client")
	}
	defer client.Close()
	req := kmspb.CreateCryptoKeyRequest{
		Parent:      config.GetKeyRing(),
		CryptoKeyId: keyId,
		CryptoKey: &kmspb.CryptoKey{
			Purpose: kmspb.CryptoKey_ENCRYPT_DECRYPT,
			VersionTemplate: &kmspb.CryptoKeyVersionTemplate{
				Algorithm: kmspb.CryptoKeyVersion_GOOGLE_SYMMETRIC_ENCRYPTION,
			},
		},
	}
	_, err = client.CreateCryptoKey(g.ctx, &req)
	if err != nil {
		return errors.Wrap(err, "failed to create crypto key")
	}
	return nil
}

func (g *GcpService) CheckIfKeyExists(keyId string) (bool, error) {
	client, err := kms.NewKeyManagementClient(g.ctx)
	if err != nil {
		return false, errors.Wrap(err, "failed to create key client")
	}
	defer client.Close()
	req := kmspb.GetCryptoKeyRequest{
		Name: config.GetKeyFullName(keyId),
	}
	_, err = client.GetCryptoKey(g.ctx, &req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to query crypto key")
	}
	return true, nil
}

func (g *GcpService) EncryptWithKMS(keyID, plaintext string) (string, error) {
	ctx := context.Background()
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to create kms client")
	}
	defer client.Close()

	// Convert the plaintext string to bytes
	plaintextBytes := []byte(plaintext)

	// Build the key name
	keyName := config.GetKeyFullName(keyID)
	// Call the API
	req := &kmspb.EncryptRequest{
		Name:      keyName,
		Plaintext: plaintextBytes,
	}

	resp, err := client.Encrypt(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to encrypt using key %s", keyName))
	}

	// Return in base64 encoding
	result := base64.StdEncoding.EncodeToString(resp.Ciphertext)

	return result, nil
}

func (g *GcpService) DecryptWithKMS(keyID, ciphertextB64 string) (string, error) {
	ctx := context.Background()
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to create kms client")
	}
	defer client.Close()

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode base64")
	}

	// Build the key name
	keyName := config.GetKeyFullName(keyID)

	// Call the API
	req := &kmspb.DecryptRequest{
		Name:       keyName,
		Ciphertext: ciphertext,
	}

	resp, err := client.Decrypt(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to decrypt using key %s", keyName))
	}

	return string(resp.Plaintext), nil
}

func (g *GcpService) GrantServiceAccountKeyRole(serviceAccountEmail string, keyId string, role string) error {
	client, err := kms.NewKeyManagementClient(g.ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create key client")
	}
	defer client.Close()
	keyName := config.GetKeyFullName(keyId)
	policy, err := client.GetIamPolicy(g.ctx, &iampb.GetIamPolicyRequest{
		Resource: keyName,
	})
	if err != nil {
		return errors.Wrap(err, "failed to get key iam policy")
	}
	policy.Bindings = append(policy.Bindings, &iampb.Binding{
		Role:    role,
		Members: []string{"serviceAccount:" + serviceAccountEmail},
	})
	res, err := client.SetIamPolicy(g.ctx, &iampb.SetIamPolicyRequest{
		Resource: keyName,
		Policy:   policy,
	})
	if err != nil {
		return errors.Wrap(err, "failed set key iam policy")
	}
	hlog.Infof("[GCPService] Bind service account %s with role %s on key %s: %v", serviceAccountEmail, role, keyId, res)
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
			IssuerUri:        config.GetIssuerUri(),
			AllowedAudiences: config.GetAllowedAudiences(),
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

func (g *GcpService) UpdateWorkloadIdentityPoolProvider(name string, imageDigest string) error {
	requestBody, err := workloadIdentityRequestBody(name, imageDigest)
	if err != nil {
		return err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("PATCH", config.GetUpdateWipProviderUrl(name), bytes.NewReader(requestBody))
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
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read http response")
	}
	hlog.Debugf("%v", string(res))
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

func convertLabelToMap(labels []*computepb.Items) map[string]string {
	res := make(map[string]string)
	for _, label := range labels {
		res[*label.Key] = *label.Value
	}
	return res
}

func covertInstanceStatus(s *string) int {
	switch *s {
	case "RUNNING":
		return INSTANCE_RUNNING
	case "TERMINATED":
		return INSTANCE_TERMINATED
	default:
		return INSTANCE_OTHER
	}
}

func (g *GcpService) ListAllInstances() ([]*Instance, error) {
	ctx := g.ctx
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create instance rest client")
	}
	defer c.Close()

	req := &computepb.ListInstancesRequest{
		Zone:    config.GetZone(),
		Project: config.GetProject(),
	}
	it := c.List(ctx, req)
	instances := make([]*Instance, 0)
	for {
		resp, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to list instances")
		}
		metadataInfo := resp.GetMetadata()
		labelMap := convertLabelToMap(metadataInfo.Items)
		instance := &Instance{
			Name:         *resp.Name,
			Status:       covertInstanceStatus(resp.Status),
			UUID:         labelMap["JOB-UUID"],
			Token:        labelMap["tee-env-USER_TOKEN"],
			CreationTime: *resp.CreationTimestamp,
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func (g *GcpService) DeleteInstance(instanceName string) error {
	projectId := config.GetProject()
	zone := config.GetZone()

	ctx := g.ctx
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create gcp instance rest client")
	}
	defer c.Close()
	req := &computepb.DeleteInstanceRequest{
		Project:  projectId,
		Zone:     zone,
		Instance: instanceName,
	}
	op, err := c.Delete(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to delete instance")
	}
	if err = op.Wait(ctx); err != nil {
		return errors.Wrap(err, "failed to wait for delete operation to complete")
	}
	return nil
}

func (g *GcpService) getTokenForStage2(string) (string, error) {
	// TODO: add authentication for statge 2
	return "", nil
}

func (g *GcpService) CreateConfidentialSpace(instanceName string, dockerImage string, stage1Token string, uuid string) error {
	stage2Token, err := g.getTokenForStage2(stage1Token)
	if err != nil {
		return err
	}
	ctx := g.ctx
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create gcp instance rest client")
	}
	defer c.Close()

	req := g.GetConfidentialSpaceInsertInstanceRequest(instanceName, dockerImage, stage2Token, uuid)

	op, err := c.Insert(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to insert instance")
	}
	if err = op.Wait(ctx); err != nil {
		return errors.Wrap(err, "failed to wait for insert operation to complete")
	}
	return nil
}

func getLogRedirectFlag() *string {
	if config.IsDebug() {
		return proto.String("true")
	}
	return proto.String("false")
}

func (g *GcpService) GetConfidentialSpaceInsertInstanceRequest(instanceName string, dockerImage string, stage2Token string, uuid string) *computepb.InsertInstanceRequest {
	cvmServiceAccountEmail := config.GetCvmServiceAccountEmail()
	zone := config.GetZone()
	machineType := config.GetMachineType()
	network := config.GetNetwork()
	subNetwork := config.GetSubNetwork()
	imageSource := config.GetTEEImageSource()
	logRedirectFlag := getLogRedirectFlag()
	diskSize := config.GetCvmDiskSize()
	instanceResource := computepb.Instance{
		ConfidentialInstanceConfig: &computepb.ConfidentialInstanceConfig{
			EnableConfidentialCompute: proto.Bool(true),
		},
		ShieldedInstanceConfig: &computepb.ShieldedInstanceConfig{
			EnableSecureBoot: proto.Bool(true),
		},
		Metadata: &computepb.Metadata{
			Items: []*computepb.Items{&computepb.Items{
				Key:   proto.String("tee-container-log-redirect"),
				Value: logRedirectFlag,
			}, &computepb.Items{
				Key:   proto.String("tee-image-reference"),
				Value: &dockerImage,
			}, &computepb.Items{
				Key:   proto.String("tee-env-USER_TOKEN"),
				Value: &stage2Token,
			}, &computepb.Items{
				Key:   proto.String("tee-env-EXECUTION_STAGE"),
				Value: proto.String("2"),
			}, &computepb.Items{
				Key:   proto.String("tee-env-DEPLOYMENT_ENV"),
				Value: proto.String(config.GetEnv()),
			}, &computepb.Items{}, &computepb.Items{
				Key:   proto.String("tee-env-PROJECT_ID"),
				Value: proto.String(config.GetProject()),
			}, &computepb.Items{}, &computepb.Items{
				Key:   proto.String("tee-env-KEY_LOCATION"),
				Value: proto.String(config.GetRegion()),
			}, &computepb.Items{
				Key:   proto.String("JOB-UUID"),
				Value: proto.String(uuid),
			},
			},
		},
		Tags: &computepb.Tags{
			Items: []string{"tee-instance"},
		},
		ServiceAccounts: []*computepb.ServiceAccount{
			&computepb.ServiceAccount{
				Email:  &cvmServiceAccountEmail,
				Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
		Name:        &instanceName,
		MachineType: &machineType,
		Disks: []*computepb.AttachedDisk{
			&computepb.AttachedDisk{
				DiskSizeGb: &diskSize,
				AutoDelete: proto.Bool(true),
				Boot:       proto.Bool(true),
				InitializeParams: &computepb.AttachedDiskInitializeParams{
					SourceImage: &imageSource,
				},
			},
		},
		Scheduling: &computepb.Scheduling{
			OnHostMaintenance: proto.String("TERMINATE"),
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			&computepb.NetworkInterface{
				AccessConfigs: []*computepb.AccessConfig{
					&computepb.AccessConfig{
						Name: proto.String("external-nat"),
						Type: proto.String("ONE_TO_ONE_NAT"),
					},
				},
				Network:    &network,
				Subnetwork: &subNetwork,
			},
		},
		CanIpForward: proto.Bool(false),
	}
	req := &computepb.InsertInstanceRequest{
		InstanceResource: &instanceResource,
		Zone:             zone,
		Project:          config.GetProject(),
	}
	return req
}

func (g *GcpService) PrepareResourcesForUser(user string) error {
	exist, err := g.CheckIfKeyExists(config.GetUserKey(user))
	if err != nil {
		return err
	}
	if !exist {
		err := g.CreateSymmetricKeys(config.GetUserKey(user))
		if err != nil {
			return err
		}
	}

	if metadata.OnGCE() {
		serviceAccountEmail, err := g.GetServiceAccountEmail()
		if err != nil {
			return err
		}
		err = g.GrantServiceAccountKeyRole(serviceAccountEmail, config.GetUserKey(user), "roles/cloudkms.cryptoKeyEncrypter")
		if err != nil {
			return err
		}
	}
	// create the workload identity pool provider for the user. The image digest set empty
	wipProvider := config.GetUserWipProvider(user)
	err = g.CreateWorkloadIdentityPoolProvider(wipProvider)
	if err != nil {
		return err
	}

	return nil
}
