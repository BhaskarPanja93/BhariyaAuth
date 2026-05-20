package mail

import (
	Logs "BhariyaAuth/processors/logs"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
)

func refreshCredentials() {
	_, _, _ = group.Do("refreshCredentials", func() (any, error) {

		Logs.RootLogger.Add(Logs.Intent, "processors/mail/main", "", "Attempting credential refresh")
		if credential == nil {
			Logs.RootLogger.Add(Logs.Error, "processors/mail/main", "", "Credential refresh skipped: credential is nil")
			return nil, nil
		}

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
