package tee_backend

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/compute/metadata"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/protobuf/proto"
)

type TEEProvider interface {
	LaunchInstance(instanceName string, image string, digest string) error
	GetInstanceStatus(instanceName string) (string, error)
	CleanUpInstance(instanceName string) error
}

type TEEProviderGCPConfidentialSpace struct {
	projectId string
	region    string
	zone      string
	env       string
	saEmail   string
	debug     bool
	ctx       context.Context
	client    *compute.InstancesClient
}

func NewTEEProviderGCPConfidentialSpace(ctx context.Context, projectId string, region string, zone string, env string) (*TEEProviderGCPConfidentialSpace, error) {
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	return &TEEProviderGCPConfidentialSpace{
		projectId: projectId,
		region:    region,
		zone:      zone,
		env:       env,
		saEmail:   fmt.Sprintf("dcr-%s-cvm-sa@%s.iam.gserviceaccount.com", env, projectId),
		// FIXME: make debug configurable
		debug:  false,
		ctx:    ctx,
		client: client,
	}, nil
}

func (c *TEEProviderGCPConfidentialSpace) GetInstanceStatus(instanceName string) (string, error) {

	req := &computepb.GetInstanceRequest{
		Project:  c.projectId,
		Zone:     c.zone,
		Instance: instanceName,
	}
	resp, err := c.client.Get(c.ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get instance: %w", err)
	}

	return *resp.Status, nil
}

func (c *TEEProviderGCPConfidentialSpace) CleanUpInstance(instanceName string) error {
	err := c.deleteWorkloadIdentityPoolProvider(instanceName)
	if err != nil {
		return err
	}

	req := &computepb.DeleteInstanceRequest{
		Project:  c.projectId,
		Zone:     c.zone,
		Instance: instanceName,
	}
	op, err := c.client.Delete(c.ctx, req)
	if err != nil {
		// instance already does not exist
		return nil
	}
	if err = op.Wait(c.ctx); err != nil {
		return fmt.Errorf("failed to wait for delete operation to complete: %w", err)
	}
	return nil
}

func (c *TEEProviderGCPConfidentialSpace) LaunchInstance(instanceName string, image string, digest string) error {

	err := c.createWorkloadIdentityPoolProvider(instanceName, digest)
	if err != nil {
		return err
	}
	err = c.createConfidentialSpace(instanceName, image)
	if err != nil {
		return err
	}

	return nil
}

func (c *TEEProviderGCPConfidentialSpace) createConfidentialSpace(instanceName string, dockerImage string) error {

	req := c.getConfidentialSpaceInsertInstanceRequest(instanceName, dockerImage)

	op, err := c.client.Insert(c.ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to insert instance")
	}
	if err = op.Wait(c.ctx); err != nil {
		return errors.Wrap(err, "failed to wait for insert operation to complete")
	}
	return nil
}

func (c *TEEProviderGCPConfidentialSpace) getConfidentialSpaceInsertInstanceRequest(instanceName string, dockerImage string) *computepb.InsertInstanceRequest {
	network := fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/global/networks/dcr-%s-network", c.projectId, c.env)
	subNetwork := fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/regions/%s/subnetworks/dcr-%s-subnetwork", c.projectId, c.region, c.env)

	imageSource := "projects/confidential-space-images/global/images/confidential-space-240200"
	logRedirectFlag := "false"
	if c.debug {
		imageSource = "projects/confidential-space-images/global/images/confidential-space-debug-240200"
		logRedirectFlag = "true"
	}

	// TODO: make machine type configurable
	machineType := fmt.Sprintf("zones/%s/machineTypes/n2d-standard-2", c.zone)
	// TODO: make disksize configurable
	diskSize := int64(50)

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
				Value: &logRedirectFlag,
			}, &computepb.Items{
				Key:   proto.String("tee-image-reference"),
				Value: &dockerImage,
			}, &computepb.Items{
				Key:   proto.String("tee-env-EXECUTION_STAGE"),
				Value: proto.String("2"),
			}, &computepb.Items{
				Key:   proto.String("tee-env-DEPLOYMENT_ENV"),
				Value: proto.String(c.env),
			}, &computepb.Items{}, &computepb.Items{
				Key:   proto.String("tee-env-PROJECT_ID"),
				Value: proto.String(c.projectId),
			}, &computepb.Items{}, &computepb.Items{
				Key:   proto.String("tee-env-KEY_LOCATION"),
				Value: proto.String(c.region),
			},
			},
		},
		Tags: &computepb.Tags{
			Items: []string{"tee-instance"},
		},
		ServiceAccounts: []*computepb.ServiceAccount{
			&computepb.ServiceAccount{
				Email:  &c.saEmail,
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
		Zone:             c.zone,
		Project:          c.projectId,
	}
	return req
}

func (c *TEEProviderGCPConfidentialSpace) createWorkloadIdentityPoolProvider(name string, digest string) error {
	// truncate name to 32 characters to avoid error
	if len(name) > 32 {
		name = name[:32]
	}
	requestBody, err := c.workloadIdentityRequestBody(name, digest)
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
	createWipProviderUrl := fmt.Sprintf(
		"https://iam.googleapis.com/v1/projects/%s/locations/global/workloadIdentityPools/dcr-%s-pool/providers?workloadIdentityPoolProviderId=%s",
		c.projectId, c.env, name)
	req, err := http.NewRequest("POST", createWipProviderUrl, bytes.NewReader(requestBody))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	token, err := c.getAccessToken()
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

func (c *TEEProviderGCPConfidentialSpace) updateWorkloadIdentityPoolProvider(name string, imageDigest string) error {
	// truncate name to 32 characters to avoid error
	if len(name) > 32 {
		name = name[:32]
	}
	requestBody, err := c.workloadIdentityRequestBody(name, imageDigest)
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
	updateWipProviderUrl := fmt.Sprintf(
		"https://iam.googleapis.com/v1/projects/%s/locations/global/workloadIdentityPools/dcr-%s-pool/providers/%s?updateMask=attributeCondition",
		c.projectId, c.env, name)
	req, err := http.NewRequest("PATCH", updateWipProviderUrl, bytes.NewReader(requestBody))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	token, err := c.getAccessToken()
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

func (c *TEEProviderGCPConfidentialSpace) deleteWorkloadIdentityPoolProvider(name string) error {
	// truncate name to 32 characters to avoid error
	if len(name) > 32 {
		name = name[:32]
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}
	client := &http.Client{Transport: tr}
	deleteWipProviderUrl := fmt.Sprintf(
		"https://iam.googleapis.com/v1/projects/%s/locations/global/workloadIdentityPools/dcr-%s-pool/providers/%s",
		c.projectId, c.env, name)
	req, err := http.NewRequest("DELETE", deleteWipProviderUrl, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	token, err := c.getAccessToken()
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

func (c *TEEProviderGCPConfidentialSpace) workloadIdentityRequestBody(name string, imageDigest string) ([]byte, error) {
	attributeCondition := fmt.Sprintf(
		"assertion.submods.container.image_digest == '%s' && '%s' in assertion.google_service_accounts && assertion.swname == 'CONFIDENTIAL_SPACE'",
		imageDigest, c.saEmail)
	if !c.debug {
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

func (c *TEEProviderGCPConfidentialSpace) getAccessToken() (string, error) {
	// Get access token
	if metadata.OnGCE() {
		res, err := metadata.GetWithContext(c.ctx, "instance/service-accounts/default/token")
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
	credentials, err := google.FindDefaultCredentials(c.ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", errors.Wrap(err, "failed to find default credential")
	}
	token, err := credentials.TokenSource.Token()
	if err != nil {
		return "", errors.Wrap(err, "failed to get token source")
	}
	return token.AccessToken, nil
}
