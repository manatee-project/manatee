package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/api/option"

	"github.com/manatee-project/manatee/pkg/config"
)

const credentialConfig = `{
	"type": "external_account",
	"audience": "//iam.googleapis.com/%s",
	"subject_token_type": "urn:ietf:params:oauth:token-type:jwt",
	"token_url": "https://sts.googleapis.com/v1/token",
	"credential_source": {
	  "file": "/run/container_launcher/attestation_verifier_claims_token"
	},
	"service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken"
}`

func requireParameter(name string, para string) {
	if para == "" {
		fmt.Printf("ERROR: %s parameter is required \n", name)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func encryptBytes(ctx context.Context, keyName string, trustedServiceAccountEmail string, wippro string, sourceData []byte) ([]byte, error) {
	cc := fmt.Sprintf(credentialConfig, wippro, trustedServiceAccountEmail)
	kmsClient, err := kms.NewKeyManagementClient(ctx, option.WithCredentialsJSON([]byte(cc)))
	if err != nil {
		return nil, fmt.Errorf("creating a new KMS client with federated credentials: %w", err)
	}

	encryptRequest := &kmspb.EncryptRequest{
		Name:      keyName,
		Plaintext: sourceData,
	}
	encryptResponse, err := kmsClient.Encrypt(ctx, encryptRequest)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt source data: %w", err)
	}

	return encryptResponse.Ciphertext, nil
}

func main() {
	err := config.InitConfig()
	if err != nil {
		fmt.Printf("ERROR: failed to init config %+v \n", err)
		return
	}

	user := flag.String("user", "", "The user who submits the job")
	inputFileName := flag.String("input", "", "The input file to be encrypted")
	outputFileName := flag.String("output", "", "The encrypted file")
	impersonationServiceAccount := flag.String("impersonation", "", "The impersonation service account it used")
	flag.Parse()
	requireParameter("user", *user)
	requireParameter("input", *inputFileName)
	requireParameter("output", *outputFileName)
	requireParameter("impersonation", *outputFileName)
	keyName := config.GetKeyFullName(config.GetUserKey(*user))
	wipProvider := config.GetWipProviderFullName(config.GetUserWipProvider(*user))

	sourceData, err := os.ReadFile(*inputFileName)
	if err != nil {
		fmt.Printf("ERROR: failed to read from input file %+v \n", err)
		return
	}

	encryptedData, err := encryptBytes(context.Background(), keyName, *impersonationServiceAccount, wipProvider, sourceData)
	if err != nil {
		fmt.Printf("ERROR: encrypt output %+v \n", err)
		return
	}
	err = os.WriteFile(*outputFileName, encryptedData, 0644)
	if err != nil {
		fmt.Printf("ERROR: failed to write output to file %+v", err)
		return
	}
}
