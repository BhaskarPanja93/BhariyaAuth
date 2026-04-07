package account

import (
	Config "BhariyaAuth/constants/config"
	ResponseModels "BhariyaAuth/models/responses"
	Stores "BhariyaAuth/stores"
	"sync"
	"time"

	"github.com/goccy/go-json"
)

// Ensures listener starts only once (thread-safe)
var listenOnce sync.Once

// pendingAccountRequests maps userID → response channel
// Used to correlate async responses with requests
var pendingAccountRequests sync.Map

// ListenAccountResponses starts a background listener that:
//   - Subscribes to Redis response channel (scoped by ServerID)
//   - Receives account details responses
//   - Routes them to the correct waiting request via channel
//
// Concurrency:
// - Uses sync.Once to ensure single listener instance
// - Uses sync.Map for concurrent request tracking
//
// Flow:
//
//	subscribe → receive message → unmarshal → lookup pending request → send response
func ListenAccountResponses() {

	listenOnce.Do(func() {

		channel := Stores.RedisClient.Subscribe(Config.CtxBG, Config.AccountDetailsResponseChannel+":"+Config.ServerID).Channel()

		go func() {
			for msg := range channel {

				var response ResponseModels.AccountDetailsResponseT

				// Ignore malformed messages
				if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
					continue
				}

				// Route response to waiting request
				if val, ok := pendingAccountRequests.Load(response.UserID); ok {

					ch := val.(chan ResponseModels.AccountDetailsResponseT)

					// Non-blocking send to avoid goroutine leak
					select {
					case ch <- response: // Send to channel
					default: // Or skip the message
					}
				}
			}
		}()
	})
}

// RequestAccountDetails performs an async request over Redis and waits for response.
//
//  1. Ensures listener is running.
//  2. Registers a response channel for given userID.
//  3. Publishes request to Redis.
//  4. Waits for response or timeout.
//
// Guarantees:
// - Request/response correlation via userID.
// - Timeout protection (prevents deadlocks).
//
// Returns:
// - Account details on success.
// - Timeout or Redis error on failure.
func RequestAccountDetails(userID int32) (ResponseModels.AccountDetailsResponseT, error) {

	ListenAccountResponses()

	responseChan := make(chan ResponseModels.AccountDetailsResponseT, 1)

	// Register request
	pendingAccountRequests.Store(userID, responseChan)
	defer pendingAccountRequests.Delete(userID)

	request := ResponseModels.AccountDetailsRequestT{
		ServerID: Config.ServerID,
		UserID:   userID,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return ResponseModels.AccountDetailsResponseT{}, err
	}

	// Publish request
	err = Stores.RedisClient.
		Publish(Config.CtxBG, Config.AccountDetailsRequestChannel, payload).
		Err()

	if err != nil {
		return ResponseModels.AccountDetailsResponseT{}, err
	}

	// Wait for response or timeout
	select {
	case res := <-responseChan:
		return res, nil

	case <-time.After(2 * time.Second):
		return ResponseModels.AccountDetailsResponseT{}, DataRequestTimedOutError
	}
}

// ServeAccountDetails listens for incoming requests and responds with user data.
//
//  1. Subscribes to request channel.
//  2. Fetches user data from DB.
//  3. Publishes response to server-specific response channel.
//
// Pattern:
// - Acts like a worker handling RPC-style requests.
//
// Flow:
//
//	receive request → query DB → marshal → publish response
func ServeAccountDetails() {

	channel := Stores.RedisClient.
		Subscribe(Config.CtxBG, Config.AccountDetailsRequestChannel).
		Channel()

	for message := range channel {

		var request ResponseModels.AccountDetailsRequestT

		// Ignore malformed requests
		if err := json.Unmarshal([]byte(message.Payload), &request); err != nil {
			continue
		}

		var response ResponseModels.AccountDetailsResponseT

		// Fetch user details from DB
		err := Stores.SQLClient.QueryRow(
			Config.CtxBG,
			"SELECT user_id, mail, name, created FROM users WHERE user_id = $1 LIMIT 1",
			request.UserID,
		).Scan(
			&response.UserID,
			&response.Email,
			&response.Name,
			&response.Created,
		)

		if err != nil {
			continue
		}

		// Serialize before sending
		payload, err := json.Marshal(response)
		if err != nil {
			continue
		}

		// Send response back to requesting server
		err = Stores.RedisClient.Publish(
			Config.CtxBG,
			Config.AccountDetailsResponseChannel+":"+request.ServerID,
			payload,
		).Err()

		if err != nil {
			continue
		}
	}
}
