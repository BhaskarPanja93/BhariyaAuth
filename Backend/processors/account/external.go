package account

import (
	Config "BhariyaAuth/constants/config"
	ResponseModels "BhariyaAuth/models/responses"
	Logs "BhariyaAuth/processors/logs"
	StringProcessor "BhariyaAuth/processors/string"
	Stores "BhariyaAuth/stores"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/bytedance/sonic"
)

const externalFileName = "processors/account/external"

var listenOnce sync.Once
var pendingAccountRequests sync.Map

func pendingRequestKey(userID int32, requestID string) string {
	if requestID != "" {
		return requestID
	}
	return "uid:" + strconv.Itoa(int(userID))
}

func ListenAccountResponses() {
	listenOnce.Do(func() {
		channel := Stores.RedisClient.Subscribe(Config.CtxBG, Config.AccountDetailsResponseChannel+":"+Config.ServerID).Channel()

		go func() {
			for message := range channel {
				var response ResponseModels.AccountDetailsResponseT
				if err := sonic.Unmarshal([]byte(message.Payload), &response); err != nil {
					Logs.RootLogger.Add(Logs.Error, externalFileName, "", "Unmarshal failed: "+err.Error())
					continue
				}

				key := pendingRequestKey(response.UserID, response.RequestID)
				val, ok := pendingAccountRequests.Load(key)
				if !ok {
					continue
				}

				select {
				case val.(chan ResponseModels.AccountDetailsResponseT) <- response:
				default:
				}
			}
		}()
	})
}

func RequestAccountDetails(userID int32) (ResponseModels.AccountDetailsResponseT, error) {
	ListenAccountResponses()

	requestID := StringProcessor.SafeString(10)
	key := pendingRequestKey(userID, requestID)
	responseChannel := make(chan ResponseModels.AccountDetailsResponseT, 1)

	pendingAccountRequests.Store(key, responseChannel)
	defer pendingAccountRequests.Delete(key)

	request := ResponseModels.AccountDetailsRequestT{
		ServerID:  Config.ServerID,
		RequestID: requestID,
		UserID:    userID,
	}

	payload, err := sonic.Marshal(request)
	if err != nil {
		return ResponseModels.AccountDetailsResponseT{}, err
	}

	err = Stores.RedisClient.
		Publish(Config.CtxBG, Config.AccountDetailsRequestChannel, payload).
		Err()
	if err != nil {
		return ResponseModels.AccountDetailsResponseT{}, err
	}

	select {
	case response := <-responseChannel:
		return response, nil
	case <-time.After(4 * time.Second):
		return ResponseModels.AccountDetailsResponseT{}, errors.New("account data request: timed out")
	}
}

func ServeAccountDetails() {
	Logs.RootLogger.Add(Logs.Intent, "main", "", "Initializing Account serving")

	channel := Stores.RedisClient.
		Subscribe(Config.CtxBG, Config.AccountDetailsRequestChannel).
		Channel()

	for message := range channel {
		var request ResponseModels.AccountDetailsRequestT
		if err := sonic.Unmarshal([]byte(message.Payload), &request); err != nil {
			Logs.RootLogger.Add(Logs.Error, externalFileName, "", "Malformed request: "+message.Payload)
			continue
		}

		var response ResponseModels.AccountDetailsResponseT
		response.RequestID = request.RequestID

		err := Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT user_id, mail, name, created FROM users WHERE user_id = $1 LIMIT 1", request.UserID).Scan(&response.UserID, &response.Email, &response.Name, &response.Created)
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, externalFileName, "", "Serve account details - SQL query: "+err.Error())
			continue
		}

		payload, err := sonic.Marshal(response)
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, externalFileName, "", "Serve account details - JSON marshal: "+request.ServerID+" "+strconv.Itoa(int(request.UserID))+" "+err.Error())
			continue
		}

		err = Stores.RedisClient.Publish(Config.CtxBG, Config.AccountDetailsResponseChannel+":"+request.ServerID, payload).Err()
		if err != nil {
			Logs.RootLogger.Add(Logs.Error, externalFileName, "", "Serve account details - Redis publish: "+request.ServerID+" "+strconv.Itoa(int(request.UserID))+" "+err.Error())
			continue
		}

		Logs.RootLogger.Add(Logs.Info, externalFileName, "", "Served account details to: "+request.ServerID+" account: "+strconv.Itoa(int(request.UserID)))
	}
}
