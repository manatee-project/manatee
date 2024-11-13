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
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	CloudProvider CloudProvider `yaml:"CloudProvider"`
}

type CloudProvider struct {
	GCP GCPConfig `yaml:"GCP"`
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
	confPath := "/usr/local/dcr_conf/config.yaml"
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
