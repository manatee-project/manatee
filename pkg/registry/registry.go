package registry

import (
	"fmt"
	"os"
)

type Registry interface {
	Url() string
	BaseImage() string
}

type GoogleDockerRegistry struct {
}

func (g *GoogleDockerRegistry) Url() string {
	projectId := os.Getenv("PROJECT_ID")
	if projectId == "" {
		panic("PROJECT_ID environment variable is not present")
	}
	env := os.Getenv("ENV")
	if env == "" {
		panic("ENV environment variable is not present")
	}

	return fmt.Sprintf("us-docker.pkg.dev/%s/dcr-%s-user-images", projectId, env)
}

func (g *GoogleDockerRegistry) BaseImage() string {
	return fmt.Sprintf("%s/data-clean-room-base:latest", g.Url())
}

type MinioDockerRegistry struct {
}

func (m *MinioDockerRegistry) Url() string {
	return "registry.kube-system.svc.cluster.local"
}

func (m *MinioDockerRegistry) BaseImage() string {
	return fmt.Sprintf("%s/dcr_tee:latest", m.Url())
}

func GetRegistry() Registry {
	registryType := os.Getenv("REGISTRY_TYPE")
	if registryType == "" {
		registryType = "GCP"
	}
	var registry Registry
	if registryType == "GCP" {
		registry = &GoogleDockerRegistry{}
	} else if registryType == "MINIO" {
		registry = &MinioDockerRegistry{}
	}
	return registry
}