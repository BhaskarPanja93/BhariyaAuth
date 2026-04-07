package mail

import (
	Secrets "BhariyaAuth/constants/secrets"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	"golang.org/x/sync/singleflight"
)

// Package-level variables for Microsoft Graph mail client and credential handling.
//
// credential: Azure client credential used for authentication.
// client: Microsoft Graph client used for sending emails.
// group: singleflight group to prevent duplicate credential refresh calls.
var (
	credential *azidentity.ClientSecretCredential
	client     *graph.GraphServiceClient
	group      singleflight.Group
)

// init initializes the Azure credential and prepares the Graph client.
//
// - Creates a client credential using tenant/client secrets.
// - Immediately initializes the Graph client via refreshCredentials.
//
// Note:
// - Errors are currently ignored (should be handled in production).
func init() {
	credential, _ = azidentity.NewClientSecretCredential(
		Secrets.MicrosoftTenantId,
		Secrets.MicrosoftClientId,
		Secrets.MicrosoftClientSecret,
		nil,
	)

	refreshCredentials()
}
