#!/bin/bash
# Copyright 2024 TikTok Pte. Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

if ! command -v minikube &> /dev/null
then
    echo "Minikube is not installed. Please install it first. https://minikube.sigs.k8s.io/docs/start/"
    exit 1
fi

env="minikube"
namespace="manatee"
dbuser="manatee"
dbpwd=$(LC_ALL=C tr -dc 'a-zA-Z0-9' </dev/urandom | head -c 12)

rm -rf terraform.tfvars
echo -e "env=\"$env\"" > terraform.tfvars
echo -e "namespace=\"$namespace\"" >> terraform.tfvars
echo -e "mysql_username=\"$dbuser\"" >> terraform.tfvars
echo -e "mysql_password=\"$dbpwd\"" >> terraform.tfvars

terraform init -reconfigure
terraform apply 

eval $(minikube docker-env)