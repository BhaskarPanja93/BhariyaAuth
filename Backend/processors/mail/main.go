package mail

import (
	Secrets "BhariyaAuth/constants/secrets"
	Logs "BhariyaAuth/processors/logs"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	"golang.org/x/sync/singleflight"
)

var (
	credential *azidentity.ClientSecretCredential
	client     *graph.GraphServiceClient
	group      singleflight.Group
)

func init() {
	var err error
	credential, err = azidentity.NewClientSecretCredential(
		Secrets.MicrosoftTenantId,
		Secrets.MicrosoftClientId,
		Secrets.MicrosoftClientSecret,
		nil,
	)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, "processors/mail/main", "", "Credential creation failed: "+err.Error())
		return
	}

	refreshCredentials()
}
