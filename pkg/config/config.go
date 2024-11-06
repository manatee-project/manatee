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
	Project   string `yaml:"Project"`
	HubBucket string `yaml:"HubBucket"`
	Zone      string `yaml:"Zone"`
	Region    string `yaml:"Region"`
	Debug     bool   `yaml:"Debug"`
	Env       string `yaml:"Env"`
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

func IsDebug() bool {
	return Conf.CloudProvider.GCP.Debug
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

func GetEnv() string {
	return Conf.CloudProvider.GCP.Env
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

func GetCloudStoragePath(file string) string {
	return fmt.Sprintf("gs://%s/%s", GetBucket(), file)
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
