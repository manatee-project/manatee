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
VAR_FILE="../../.env"
if [ ! -f "$VAR_FILE" ]; then
    echo "Error: Variables file does not exist."
    exit 1
fi

VAR_FILE=$(realpath $VAR_FILE)
source $VAR_FILE

zone=$region-a

PKG_SOURCE_CODE="../../pkg"
if [ ! -d "$PKG_SOURCE_CODE" ]; then
    echo "Error: Pkg soruce codes don't exist."
    exit 1
fi

rm -rf github.com
mkdir -p github.com/manatee-project/manatee/
cp -r $PKG_SOURCE_CODE github.com/manatee-project/manatee/pkg@v0.0.1

cp -r ../conf ./
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/ENV/$env/g" conf/config.yaml
    sed -i '' "s/Project: \"\$PROJECTID\"/Project: \"$project_id\"/g" conf/config.yaml
    sed -i '' "s/ProjectNumber: \$PROJECTNUMBER/ProjectNumber: $project_number/g" conf/config.yaml
    sed -i '' "s/Region: \"\$REGION\"/Region: \"$region\"/g" conf/config.yaml
    sed -i '' "s/Zone: \"\$ZONE\"/Zone: \"$zone\"/g" conf/config.yaml
else
    sed -i "s/ENV/$env/g" conf/config.yaml
    sed -i "s/Project: \"\$PROJECTID\"/Project: \"$project_id\"/g" conf/config.yaml
    sed -i "s/ProjectNumber: \$PROJECTNUMBER/ProjectNumber: $project_number/g" conf/config.yaml
    sed -i "s/Region: \"\$REGION\"/Region: \"$region\"/g" conf/config.yaml
    sed -i "s/Zone: \"\$ZONE\"/Zone: \"$zone\"/g" conf/config.yaml
fi

# base tee image will be pushed to "dcr-${var.env}-user-images"
user_docker_repo="dcr-$env-user-images"
IMAGE_URL="us-docker.pkg.dev/${project_id}/${user_docker_repo}/data-clean-room-base"
TAG="latest"
docker build --platform linux/amd64 --tag "$IMAGE_URL:$TAG" .
docker push "$IMAGE_URL:$TAG"
