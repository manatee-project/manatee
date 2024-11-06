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

package cloud

import (
	"context"
	"io"
)

type CloudProvider interface {
	// cloud storage
	DownloadFile(remoteSrcPath string, localDestPath string) error
	ListFiles(remoteDir string) ([]string, error)
	GetFileSize(remotePath string) (int64, error)
	GetFilebyChunk(remotePath string, offset int64, chunkSize int64) ([]byte, error)
	DeleteFile(remotePath string) error
	UploadFile(fileReader io.Reader, remotePath string, compress bool) error
	// compute engine
	GetServiceAccountEmail() (string, error)
}

func GetCloudProvider(ctx context.Context) CloudProvider {
	return NewGcpService(ctx)
}
