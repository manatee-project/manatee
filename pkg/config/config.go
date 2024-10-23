// Copyright 2024 TikTok Pte. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/manatee-project/manatee/pkg/utils"
)

type Config struct {
	CloudProvider CloudProvider `yaml:"CloudProvider"`
	Cluster       Cluster       `yaml:"Cluster"`
}

type CloudProvider struct {
	GCP GCPConfig `yaml:"GCP"`
}

type Cluster struct {
	PodServiceAccount string `yaml:"PodServiceAccount"`
}

type GCPConfig struct {
	Project              string   `yaml:"Project"`
	ProjectNumber        uint64   `yaml:"ProjectNumber"`
	Repository           string   `yaml:"Repository"`
	HubBucket            string   `yaml:"HubBucket"`
	CvmServiceAccount    string   `yaml:"CvmServiceAccount"`
	Zone                 string   `yaml:"Zone"`
	Region               string   `yaml:"Region"`
	Cpus                 int      `yaml:"CPUs"`
	DiskSize             int      `yaml:"DiskSize"`
	Debug                bool     `yaml:"Debug"`
	KeyRing              string   `yaml:"KeyRing"`
	WorkloadIdentityPool string   `yaml:"WorkloadIdentityPool"`
	IssuerUri            string   `yaml:"IssuerUri"`
	AllowedAudiences     []string `yaml:"AllowedAudiences"`
	Network              string   `yaml:"Network"`
	Subnetwork           string   `yaml:"Subnetwork"`
	Env                  string   `yaml:"Env"`
}

type APIConfig struct {
	UseAuth bool `yaml:"UseAuth"`
}

var Conf Config

func InitConfig() error {
	viper.SetConfigType("yaml")
	confPath := utils.GetConfigFile()
	hlog.Infof("[Config] Config file: %s", confPath)
	viper.SetConfigFile(confPath)
	if err := viper.ReadInConfig(); err != nil {
		return errors.Wrap(err, "failed to read config")
	}
	if err := viper.Unmarshal(&Conf); err != nil {
		return errors.Wrap(err, "failed to unmarshal config")
	}
	hlog.Infof("[Config] Conf.CloudProvider: %#v", Conf.CloudProvider)
	return nil
}

func GetBucket() string {
	return Conf.CloudProvider.GCP.HubBucket
}

func GetKeyRing() string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", Conf.CloudProvider.GCP.Project, Conf.CloudProvider.GCP.Region, Conf.CloudProvider.GCP.KeyRing)
}

func GetKeyFullName(keyName string) string {
	return fmt.Sprintf("%s/cryptoKeys/%s", GetKeyRing(), keyName)
}

func getServiceAccountEmail(serviceAccount string, project string) string {
	return fmt.Sprintf("%s@%s.iam.gserviceaccount.com", serviceAccount, project)
}

func GetCvmServiceAccountEmail() string {
	return getServiceAccountEmail(Conf.CloudProvider.GCP.CvmServiceAccount, Conf.CloudProvider.GCP.Project)
}

func IsDebug() bool {
	return Conf.CloudProvider.GCP.Debug
}

func GetIssuerUri() string {
	return Conf.CloudProvider.GCP.IssuerUri
}

func GetAllowedAudiences() []string {
	return Conf.CloudProvider.GCP.AllowedAudiences
}

func GetCreateWipProviderUrl(provider string) string {
	return fmt.Sprintf("https://iam.googleapis.com/v1/projects/%s/locations/global/workloadIdentityPools/%s/providers?workloadIdentityPoolProviderId=%s", Conf.CloudProvider.GCP.Project, Conf.CloudProvider.GCP.WorkloadIdentityPool, provider)
}

func GetUpdateWipProviderUrl(provider string) string {
	return fmt.Sprintf("https://iam.googleapis.com/v1/projects/%s/locations/global/workloadIdentityPools/%s/providers/%s?updateMask=attributeCondition", Conf.CloudProvider.GCP.Project, Conf.CloudProvider.GCP.WorkloadIdentityPool, provider)
}

func GetUserWipProvider(user string) string {
	name := fmt.Sprintf("%s-tee-provider", user)
	if len(name) > 32 {
		return name[:32]
	}
	return name
}

func GetZone() string {
	return Conf.CloudProvider.GCP.Zone
}

func GetRegion() string {
	return Conf.CloudProvider.GCP.Region
}

func GetProject() string {
	return Conf.CloudProvider.GCP.Project
}

func GetMachineType() string {
	return fmt.Sprintf("zones/%s/machineTypes/n2d-standard-%d", GetZone(), Conf.CloudProvider.GCP.Cpus)
}

func GetNetwork() string {
	return fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/global/networks/%s", GetProject(), Conf.CloudProvider.GCP.Network)
}

func GetSubNetwork() string {
	return fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/regions/%s/subnetworks/%s", GetProject(), GetRegion(), Conf.CloudProvider.GCP.Subnetwork)
}

func GetEnv() string {
	return Conf.CloudProvider.GCP.Env
}

func GetTEEImageSource() string {
	if Conf.CloudProvider.GCP.Debug {
		return "projects/confidential-space-images/global/images/confidential-space-debug-240200"
	}
	return "projects/confidential-space-images/global/images/confidential-space-240200"
}

func GetCvmDiskSize() int64 {
	return int64(Conf.CloudProvider.GCP.DiskSize)
}

func GetWipProviderFullName(provider string) string {
	return fmt.Sprintf("projects/%v/locations/global/workloadIdentityPools/%s/providers/%s", Conf.CloudProvider.GCP.ProjectNumber, Conf.CloudProvider.GCP.WorkloadIdentityPool, provider)
}

func GetJobOutputFilename(UUID string, originName string) string {
	if len(UUID) >= 8 {
		return fmt.Sprintf("out-%s-%s", UUID[:8], originName)
	}
	return fmt.Sprintf("out-%s-%s", UUID, originName)
}

func GetEncryptedJobOutputFilename(UUID string, originName string) string {
	return fmt.Sprintf("enc-%s-%s", UUID[:8], originName)
}

func GetEncryptedJobOutputPath(creator, UUID, originName string) string {
	return fmt.Sprintf("%s/output/%s", creator, GetEncryptedJobOutputFilename(UUID, originName))
}

func GetJobOutputPath(creator string, UUID string, originName string) string {
	return fmt.Sprintf("%s/output/%s", creator, GetJobOutputFilename(UUID, originName))
}

func GetCustomTokenPath(creator string, UUID string) string {
	return fmt.Sprintf("%s/output/%s-token", creator, UUID)
}

func GetBaseDockerImage() string {
	return fmt.Sprintf("us-docker.pkg.dev/%s/%s/%s:latest", GetProject(), Conf.CloudProvider.GCP.Repository, "data-clean-room-base")
}

func GetCloudStoragePath(file string) string {
	return fmt.Sprintf("gs://%s/%s", GetBucket(), file)
}

func GetJobDockerImageName(creator, UUID string) string {
	return fmt.Sprintf("%s-%s", creator, UUID)
}

func GetJobDockerImageFull(creator string, UUID string) string {
	return fmt.Sprintf("us-docker.pkg.dev/%s/%s/%s:latest", GetProject(), Conf.CloudProvider.GCP.Repository, GetJobDockerImageName(creator, UUID))
}

func GetUserWorkspaceFile(creator string) string {
	return fmt.Sprintf("%s.tar.gz", GetUserWorkSpaceDir(creator))
}

func GetUserWorkSpaceDir(creator string) string {
	return fmt.Sprintf("%s-workspace", creator)
}

func GetUserWorkSpacePath(creator string) string {
	return fmt.Sprintf("%s/%s", creator, GetUserWorkspaceFile(creator))
}

func GetBuildContextFileName(UUID string) string {
	return fmt.Sprintf("context-%s.tar.gz", UUID)
}

func GetBuildContextPath(creator, UUID string) string {
	return fmt.Sprintf("%s/%s", creator, GetBuildContextFileName(UUID))
}

func GetUserKey(user string) string {
	return fmt.Sprintf("%s-key", user)
}

func GetK8sPodServiceAccount() string {
	return Conf.Cluster.PodServiceAccount
}

func GetInstanceName(creator string, UUID string) string {
	return fmt.Sprintf("%s-%s", creator, UUID[:8])
}
