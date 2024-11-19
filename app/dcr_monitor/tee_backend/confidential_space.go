package tee_backend

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type TEEProvider interface {
	LaunchInstance(instanceName string, image string, digest string, extraEnvs map[string]string) error
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

func (c *TEEProviderGCPConfidentialSpace) LaunchInstance(instanceName string, image string, digest string, extraEnvs map[string]string) error {
	err := c.createConfidentialSpace(instanceName, image, extraEnvs)
	if err != nil {
		return err
	}

	return nil
}

func (c *TEEProviderGCPConfidentialSpace) createConfidentialSpace(instanceName string, dockerImage string, extraEnvs map[string]string) error {

	req := c.getConfidentialSpaceInsertInstanceRequest(instanceName, dockerImage, extraEnvs)

	op, err := c.client.Insert(c.ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to insert instance")
	}
	if err = op.Wait(c.ctx); err != nil {
		return errors.Wrap(err, "failed to wait for insert operation to complete")
	}
	return nil
}

func (c *TEEProviderGCPConfidentialSpace) getConfidentialSpaceInsertInstanceRequest(instanceName string, dockerImage string, extraEnvs map[string]string) *computepb.InsertInstanceRequest {
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

	metadataItems := []*computepb.Items{&computepb.Items{
		Key:   proto.String("tee-container-log-redirect"),
		Value: &logRedirectFlag,
	}, &computepb.Items{
		Key:   proto.String("tee-image-reference"),
		Value: &dockerImage,
	},
	}

	for key, value := range extraEnvs {
		metadataItems = append(metadataItems, &computepb.Items{
			Key:   proto.String(fmt.Sprintf("tee-env-%s", key)),
			Value: proto.String(value),
		})
	}

	instanceResource := computepb.Instance{
		ConfidentialInstanceConfig: &computepb.ConfidentialInstanceConfig{
			EnableConfidentialCompute: proto.Bool(true),
		},
		ShieldedInstanceConfig: &computepb.ShieldedInstanceConfig{
			EnableSecureBoot: proto.Bool(true),
		},
		Metadata: &computepb.Metadata{
			Items: metadataItems,
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
