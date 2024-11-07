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
    echo -e "Error: the namespace parameter is missing, please run the script like ./push_image.sh --namespace=xxx"
    exit
fi

source env.bzl

bazel run //:push_dcr_api_image -- --repository "us-docker.pkg.dev/$project_id/dcr-$env-$namespace-images/data-clean-room-api"
bazel run //:push_dcr_monitor_image -- --repository "us-docker.pkg.dev/$project_id/dcr-$env-$namespace-images/data-clean-room-monitor"
bazel run //:push_jupyterlab_image -- --repository "us-docker.pkg.dev/$project_id/dcr-$env-$namespace-images/scipy-notebook-with-dcr"
bazel run //:push_dcr_tee_image 
