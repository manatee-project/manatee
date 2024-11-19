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
	"os"
	"strconv"

	"github.com/cloudwego/hertz/pkg/common/hlog"
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
	env := os.Getenv("ENV")
	projectId := os.Getenv("PROJECT_ID")
	region := os.Getenv("REGION")
	zone := os.Getenv("ZONE")
	debug := os.Getenv("DEBUG")
	hlog.Infof("env %s project id %s region %s zone %s debug %s", env, projectId, region, zone, debug)
	d, err := strconv.ParseBool(debug)
	if err != nil {
		d = false
	}
	Conf := Config{
		CloudProvider: CloudProvider{
			GCP: GCPConfig{
				Project:   projectId,
				HubBucket: fmt.Sprintf("dcr-%s-debug", env),
				Zone:      zone,
				Region:    region,
				Debug:     d,
				Env:       env,
			},
		},
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
