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
set -e

for arg in "$@"
do
    case $arg in
        --namespace=*)
        # If we find an argument --namespace=something, split the string into a name/value array.
        IFS='=' read -ra NAMESPACE <<< "$arg"
        # Assign the second element of the array (the value of the --namespace argument) to our variable.
        namespace="${NAMESPACE[1]}"
        ;;
        --database-user=*)
        IFS='=' read -ra DBUSER <<< "$arg"
        dbuser="${DBUSER[1]}"
        ;;
        --database-password=*)
        IFS='=' read -ra DBPWD <<< "$arg"
        dbpwd="${DBPWD[1]}"
    esac
done


if [ -z "$namespace" ]; then
    echo -e "Error: the namespace parameter is missing, please run the script like ./apply.sh --namespace=xxx"
    exit
fi

if [ -z "$dbuser" ]; then
    dbuser=$namespace
    echo -e "\033[1;33mWarning: the database-user parameters doesn't exist using default database user: ${namespace}\033[0m"
    
fi

if [ -z "$dbpwd" ]; then
    dbpwd=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
    echo -e "\033[1;33mWarning: the database-password parameters doesn't exist using random database password: ${dbpwd}\033[0m"
fi

# Check if gcloud is installed
if ! [ -x "$(command -v gcloud)" ]; then
	echo "Error: gcloud is not installed." >&2
	exit 1
fi

# Check if gcloud logged in
if ! gcloud auth list | grep -q 'ACTIVE'; then
	echo "Error: No active gcloud account found." >&2
	exit 1
fi

# check whether variables has been set
VAR_FILE="../../env.bzl"
if [ ! -f "$VAR_FILE" ]; then
    echo "Error: Variables file does not exist."
    exit 1
fi
VAR_FILE=$(realpath $VAR_FILE)
source $VAR_FILE

zone=$region-a
# get kubernete cluster credentials
gcloud container clusters get-credentials dcr-$env-cluster --zone $zone --project $project_id

cp $VAR_FILE terraform.tfvars

echo -e "\nnamespace=\"$namespace\"" >> terraform.tfvars
echo -e "mysql_username=\"$dbuser\"" >> terraform.tfvars
echo -e "mysql_password=\"$dbpwd\"" >> terraform.tfvars
terraform init -reconfigure -backend-config="bucket=dcr-tf-state-$env"  -backend-config="prefix=$namespace" 

terraform apply
