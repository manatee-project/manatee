package tee_backend

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
)

type TEEProvider interface {
	// LaunchInstance(ctx context.Context, imageName string) (string, error)
	GetInstanceStatus(instanceName string) (string, error)
	CleanUpInstance(instanceName string) error
}

type TEEProviderGCPConfidentialSpace struct {
	projectId string
	zone      string
	ctx       context.Context
	client    *compute.InstancesClient
}

func NewTEEProviderGCPConfidentialSpace(ctx context.Context, projectId string, zone string) (*TEEProviderGCPConfidentialSpace, error) {
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	return &TEEProviderGCPConfidentialSpace{
		projectId: projectId,
		zone:      zone,
		ctx:       ctx,
		client:    client,
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
