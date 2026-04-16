package mail

import (
	Logs "BhariyaAuth/processors/logs"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
)

// refreshCredentials refreshes the Microsoft Graph client instance.
//
// - Uses singleflight to ensure only one refresh happens concurrently.
// - Recreates the GraphServiceClient using existing credentials.
// - Updates global client reference if successful.
//
// Concurrency Model:
// - Prevents duplicate refresh calls under high contention.
//
// Returns:
// - No direct return (updates global state).
func refreshCredentials() {
	_, _, _ = group.Do("refreshCredentials", func() (any, error) {

		Logs.RootLogger.Add(Logs.Intent, "processors/mail/main", "", "Credentials refreshing")

		c, err := graph.NewGraphServiceClientWithCredentials(
			credential,
			[]string{"https://graph.microsoft.com/.default"},
		)

		if err == nil {
			Logs.RootLogger.Add(Logs.Info, "processors/mail/main", "", "Credentials refreshed")
			client = c
		}

		return nil, err
	})
}
