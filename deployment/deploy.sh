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
    esac
done


if [ -z "$namespace" ]; then
    echo "Error: the namespace parameter is required, run the script again like ./apply.sh --namespace="
    exit 1
fi

deploy_service() {
    app=$1
    pushd $app
    ./deploy.sh $2
    popd
}

deploy_service data-clean-room $namespace
deploy_service jupyterhub $namespace
