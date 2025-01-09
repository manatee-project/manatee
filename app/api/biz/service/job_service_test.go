package service

import (
	"context"
	"os"
	"strings"
	"testing"
)

var expectedDockerfile1 string = `ARG BASE_IMAGE
FROM $BASE_IMAGE
ARG OUTPUTPATH
ARG JUPYTER_FILENAME
ARG USER_WORKSPACE
ARG CUSTOMTOKEN_CLOUDSTORAGE_PATH 

ENV OUTPUTPATH=$OUTPUTPATH
ENV JUPYTER_FILENAME=$JUPYTER_FILENAME
ENV CUSTOMTOKEN_CLOUDSTORAGE_PATH=$CUSTOMTOKEN_CLOUDSTORAGE_PATH

WORKDIR /home/jovyan
COPY $USER_WORKSPACE/* ./


ENTRYPOINT jupyter nbconvert --execute --to notebook --inplace $JUPYTER_FILENAME --ExecutePreprocessor.timeout=-1 --allow-errors \
    && hash=$(md5sum $JUPYTER_FILENAME | awk '{ print $1 }') \
    && ./gscp $JUPYTER_FILENAME $OUTPUTPATH \
    && ./gen_custom_token --nonce $hash \
    && ./gscp custom_token $CUSTOMTOKEN_CLOUDSTORAGE_PATH
`

func TestGenerateDockerfile(t *testing.T) {
	os.Setenv("STORAGE_TYPE", "MOCK")
	os.Setenv("ENV", "minikube")
	js := NewJobService(context.Background())

	content := js.generateDockerfile([]string{})
	if strings.Contains(content, `LABEL "tee.launch_policy.allow_env_override"`) {
		t.Errorf("Dockerfile contains wrong allow_env_override policy")
	}
	content = js.generateDockerfile([]string{"USER_TOKEN"})
	if !strings.Contains(content, `LABEL "tee.launch_policy.allow_env_override"="USER_TOKEN"`) {
		t.Errorf("Dockerfile does not contain correct allow_env_override policy")
	}
	content = js.generateDockerfile([]string{"USER_TOKEN", "CUSTOM_ENV_VAR", "BREAKPOINT"})
	if !strings.Contains(content, `LABEL "tee.launch_policy.allow_env_override"="USER_TOKEN,CUSTOM_ENV_VAR,BREAKPOINT"`) {
		t.Errorf("Dockerfile does not contain correct allow_env_override policy")
	}
}
