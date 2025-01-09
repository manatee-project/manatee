// Copyright 2024 TikTok Pte. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const TikTokAudience = "https://research.tiktok.com/"
const TokenFilename = "custom_token"

type CustomToken struct {
	Audience  string   `json:"audience"`
	Nonces    []string `json:"nonces"` // each nonce must be min 64bits
	TokenType string   `json:"token_type"`
}

func GcsCustomAttestationToken(nonce string) ([]byte, error) {
	request := CustomToken{
		Audience:  TikTokAudience,
		Nonces:    []string{nonce},
		TokenType: "OIDC",
	}
	httpClient := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/run/container_launcher/teeserver.sock")
			},
		},
	}
	customJSON, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}
	url := "http://localhost/v1/token"
	resp, err := httpClient.Post(url, "application/json", strings.NewReader(string(customJSON)))
	if err != nil {
		return nil, errors.Wrap(err, "faile to get custom token")
	}
	defer resp.Body.Close()
	tokenbytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "faile to read from response")
	}

	return tokenbytes, nil
}

func generateCustomAttestationToken(nonce string) ([]byte, error) {
	if os.Getenv("TEE_BACKEND") == "MOCK" {
		return []byte(fmt.Sprintf("mock tee token with nonce %s", nonce)), nil
	} else {
		return GcsCustomAttestationToken(nonce)
	}
}

func requireParameter(name string, para string) {
	if para == "" {
		fmt.Printf("ERROR: %s parameter is required \n", name)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	nonce := flag.String("nonce", "", "The nonce to generate custom token")
	flag.Parse()
	requireParameter("nonce", *nonce)
	customToken, err := generateCustomAttestationToken(*nonce)
	if err != nil {
		fmt.Printf("ERROR: failed to generate custom token %+v \n", err)
		panic(err)
	}

	err = os.WriteFile(TokenFilename, customToken, 0644)
	if err != nil {
		fmt.Printf("ERROR: failed to write custom token to file %+v \n", err)
		log.Fatal(err)
	}
}
