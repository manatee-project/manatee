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
	return fmt.Sprintf("%s/manatee-executor-base:latest", g.Url())
}

type MinikubeDockerRegistry struct {
}

func (m *MinikubeDockerRegistry) Url() string {
	return "registry.kube-system.svc.cluster.local"
}

func (m *MinikubeDockerRegistry) BaseImage() string {
	return fmt.Sprintf("%s/executor:latest", m.Url())
}

func GetRegistry() Registry {
	registryType := os.Getenv("REGISTRY_TYPE")
	if registryType == "" {
		registryType = "GCP"
	}
	var registry Registry
	if registryType == "GCP" {
		registry = &GoogleDockerRegistry{}
	} else if registryType == "MINIKUBE" {
		registry = &MinikubeDockerRegistry{}
	}
	return registry
}
