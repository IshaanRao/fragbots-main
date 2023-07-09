package routes

import (
	"errors"
	"fragbotsbackend/constants"
	"fragbotsbackend/database"
	"fragbotsbackend/logging"
	"fragbotsbackend/servers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

// Requests

type CreateBotRequest struct {
	Stage    int    `json:"stage" form:"stage"`
	Email    string `json:"email,omitempty" form:"email"`
	Password string `json:"password,omitempty" form:"password"`
	UserCode string `json:"userCode,omitempty" form:"userCode"`
}

type StopBotRequest struct {
	HardStop bool `json:"hardStop" form:"hardStop"`
}

// Response Types

type AccountDataResponse struct {
	Token   string `json:"token"`
	Profile struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"profile"`
}

// Misc Types

type MsAuthChannelData struct {
	Channel chan *constants.MSauth
	BotData *constants.BotInfo
}

var authChannels = make(map[string]MsAuthChannelData)

func getVerifiedOneBotInfo() *constants.BotInfo {
	return &constants.BotInfo{
		BotId:   "Verified1",
		BotType: constants.Verified,
		DiscInfo: &constants.DiscordInfo{
			LogWebhook:     constants.VerifiedOneLogWebhook,
			ConsoleWebhook: constants.VerifiedOneConsoleWebhook,
		},
		AccInfo: &constants.AccountInfo{},
	}
}

func getVerifiedTwoBotInfo() *constants.BotInfo {
	return &constants.BotInfo{
		BotId:   "Verified2",
		BotType: constants.Verified,
		DiscInfo: &constants.DiscordInfo{
			LogWebhook:     constants.VerifiedTwoLogWebhook,
			ConsoleWebhook: constants.VerifiedTwoConsoleWebhook,
		},
		AccInfo: &constants.AccountInfo{},
	}
}

func getWhitelistedBotInfo() *constants.BotInfo {
	return &constants.BotInfo{
		BotId:   "Whitelisted",
		BotType: constants.Whitelisted,
		DiscInfo: &constants.DiscordInfo{
			LogWebhook:     constants.WhitelistedLogWebhook,
			ConsoleWebhook: constants.WhitelistedConsoleWebhook,
		},
		AccInfo: &constants.AccountInfo{},
	}
}

func getActiveBotInfo() *constants.BotInfo {
	return &constants.BotInfo{
		BotId:   "Active",
		BotType: constants.Active,
		DiscInfo: &constants.DiscordInfo{
			LogWebhook:     constants.ActiveLogWebhook,
			ConsoleWebhook: constants.ActiveConsoleWebhook,
		},
		AccInfo: &constants.AccountInfo{},
	}
}

func getExclusiveBotInfo() *constants.BotInfo {
	return &constants.BotInfo{
		BotId:   "Exclusive",
		BotType: constants.Exclusive,
		DiscInfo: &constants.DiscordInfo{
			LogWebhook:     constants.ExclusiveLogWebhook,
			ConsoleWebhook: constants.ExclusiveConsoleWebhook,
		},
		AccInfo: &constants.AccountInfo{},
	}
}

func getPriorityBotInfo() *constants.BotInfo {
	return &constants.BotInfo{
		BotId:   "Priority",
		BotType: constants.Priority,
		DiscInfo: &constants.DiscordInfo{
			LogWebhook:     constants.PriorityLogWebhook,
			ConsoleWebhook: constants.PriorityConsoleWebhook,
		},
		AccInfo: &constants.AccountInfo{},
	}
}

func getBotInfo(botId string) (*constants.BotInfo, error) {
	var botInfo *constants.BotInfo = nil
	switch botId {
	case "Priority":
		botInfo = getPriorityBotInfo()
	case "Exclusive":
		botInfo = getExclusiveBotInfo()
	case "Active":
		botInfo = getActiveBotInfo()
	case "Whitelisted":
		botInfo = getWhitelistedBotInfo()
	case "Verified1":
		botInfo = getVerifiedOneBotInfo()
	case "Verified2":
		botInfo = getVerifiedTwoBotInfo()
	}
	if botInfo == nil {
		return nil, errors.New("invalid id")
	}
	return botInfo, nil
}

func getBotInfoFromDb(botId string) (*constants.BotInfo, error) {
	var botInfo constants.BotInfo
	err := database.GetDocument("accounts", bson.D{{"botId", botId}}, &botInfo)
	if err != nil {
		return nil, err
	}
	return &botInfo, nil
}

// GetBot gives bot data to fragbots for startup, refreshes login tokens
func GetBot(c *gin.Context) {
	id := c.Param("botid")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Missing BotID"})
		return
	}

	botInfo, err := getBotInfoFromDb(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid BotID"})
		return
	}

	var MSa = botInfo.AccInfo.AuthData
	err = CheckRefreshMS(&MSa)
	if err != nil {
		logging.LogWarn("Failed to check refresh token: " + err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Failed to check refresh token"})
		return
	}

	if MSa.AccessToken != botInfo.AccInfo.AuthData.AccessToken {
		botInfo.AccInfo.AuthData = MSa
	}

	credentials, err := GetMCcredentials(botInfo.AccInfo.AuthData)
	if err != nil {
		logging.LogWarn("Error occurred while getting credentials: " + err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error occured while checking data"})
		return
	}
	botInfo.AccInfo.UUID = credentials.UUID
	botInfo.AccInfo.Username = credentials.Username
	botInfo.AccInfo.AccessToken = credentials.AccessToken
	logging.Log("Successfully refreshed bot data")

	c.IndentedJSON(http.StatusOK, gin.H{"botInfo": botInfo})
}

func RestartBot(c *gin.Context) {
	botId := c.Param("botid")

	botInfo, err := getBotInfoFromDb(botId)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid BotID"})
		return
	}

	if botInfo.ServerId == "" {
		c.IndentedJSON(http.StatusConflict, gin.H{"error": "bot server needs to be started (use startbot)"})
		return
	}

	if botInfo.Running {
		_ = servers.StopFragbotService(botId)
	}

	err = servers.RunFragbotsService(botId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to start fragbot: " + err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"success": true})

}

// StopBot stops fragbot service and removes the aws server if specified in request
func StopBot(c *gin.Context) {
	botId := c.Param("botid")
	var request StopBotRequest
	err := c.Bind(&request)
	if botId == "" || err != nil {
		if err != nil {
			logging.LogWarn(err.Error())
		}
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
	}

	botInfo, err := getBotInfoFromDb(botId)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid BotID"})
		return
	}

	if botInfo.Running {
		_ = servers.StopFragbotService(botId)
		_ = database.UpdateDocument("accounts", bson.D{{"botId", botId}}, bson.D{{"running", false}})
	}

	if request.HardStop && botInfo.ServerId != "" {
		err = servers.DeleteInstance(botInfo.ServerId)
		if err != nil {
			logging.LogWarn(err.Error())
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Failed to delete instance"})
			return
		}
		servers.RemoveFragbotNode(botId)
		err = database.UpdateDocumentDelField("accounts", bson.D{{"botId", botId}}, bson.D{{"serverId", nil}})
		if err != nil {
			logging.LogWarn("Failed to remove server id: " + err.Error())
		}
	}

	c.IndentedJSON(http.StatusOK, gin.H{"success": true})
}

func DeleteBot(c *gin.Context) {
	botId := c.Param("botid")

	if botId == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid bot id"})
		return
	}

	botInfo, err := getBotInfoFromDb(botId)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid BotID"})
		return
	}

	if botInfo.Running {
		_ = servers.StopFragbotService(botId)
		_ = database.UpdateDocument("accounts", bson.D{{"botId", botId}}, bson.D{{"running", false}})
	}

	if botInfo.ServerId != "" {
		err = servers.DeleteInstance(botInfo.ServerId)
		if err != nil {
			logging.LogWarn(err.Error())
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Failed to delete instance"})
			return
		}
		servers.RemoveFragbotNode(botId)
		err = database.UpdateDocumentDelField("accounts", bson.D{{"botId", botId}}, bson.D{{"serverId", nil}})
		if err != nil {
			logging.LogWarn("Failed to remove server id: " + err.Error())
		}

	}

	err = database.UpdateDocument("accounts", bson.D{{"botId", botId}}, bson.D{{"botId", "archive_" + botId}})
	if err != nil {
		logging.LogWarn("Failed to archive document err, " + err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document when archiving"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"success": true})

}

// StartBot either creates aws server or starts fragbot service on server
func StartBot(c *gin.Context) {
	botId := c.Param("botid")
	if botId == "" {
		logging.LogWarn("Put bot request failed invalid request body")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	var botInfo constants.BotInfo
	err := database.GetDocument("accounts", bson.D{{"botId", botId}}, &botInfo)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid BotID"})
		return
	}
	if botInfo.Running {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Bot currently active"})
		return
	}
	if botInfo.ServerId == "" {
		id, err := servers.MakeFragBotServer(botInfo.BotId)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "error occurred while starting server"})
			return
		}
		err = database.UpdateDocument("accounts", bson.D{{"botId", botId}}, bson.D{{"serverId", id}})
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "error occurred while updating serverid in db PLEASE TELL PRINCE"})
			return
		}
	} else {
		err := servers.RunFragbotsService(botId)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "error occurred while starting service"})
			return
		}
	}
	c.IndentedJSON(http.StatusOK, gin.H{"success": true})
}

// CreateBot has two stages to create a fragbot
// stage 1 uses credentials give user a link that allows them to auth mc acct
// stage two creates bot in db and starts aws server
func CreateBot(c *gin.Context) {
	botId := c.Param("botid")
	var request CreateBotRequest
	err := c.Bind(&request)

	if botId == "" || err != nil || request.Stage == 0 {
		logging.LogWarn("Post bot request failed invalid request body")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	info, err := getBotInfo(botId)
	if err != nil {
		logging.LogWarn("Create bot request failed: " + err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
		return
	}

	if request.Stage == 1 {
		if request.Password == "" || request.Email == "" {
			logging.LogWarn("Post bot request failed invalid request body (Stage 1)")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if (database.DocumentExists("accounts", bson.D{{"botId", botId}})) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Bot already exists with that id"})
			return
		}

		// Device flow with account
		msAuthData, authChannel, err := AuthMSdevice()
		if err != nil {
			logging.LogWarn(err.Error())
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}
		info.AccInfo.Email = request.Email
		info.AccInfo.Password = request.Password
		authChannels[msAuthData.UserCode] = MsAuthChannelData{
			Channel: authChannel,
			BotData: info,
		}
		msAuthData.Email = request.Email
		msAuthData.Password = request.Password
		c.IndentedJSON(http.StatusOK, gin.H{"msAuthInfo": msAuthData})
		return

	} else if request.Stage == 2 {
		if request.UserCode == "" {
			logging.LogWarn("Post bot request failed invalid request body (Stage 1)")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if (database.DocumentExists("accounts", bson.D{{"botId", botId}})) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Bot already exists with that id"})
			return
		}

		msData, ok := authChannels[request.UserCode]
		if !ok {
			logging.LogWarn("Invalid User Code")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid User Code"})
			return
		}

		authData := <-msData.Channel
		if authData == nil {
			logging.LogWarn("Error occurred while checking data most likely timeout")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error occured while checking data"})
			return
		}

		msData.BotData.AccInfo.AuthData = *authData

		credentials, err := GetMCcredentials(*authData)
		if err != nil {
			logging.LogWarn("Error occurred while getting credentials: " + err.Error())
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error occured while checking data"})
			return
		}
		delete(authChannels, request.UserCode)

		botInfo := *msData.BotData
		botInfo.AccInfo.UUID = credentials.UUID
		botInfo.AccInfo.Username = credentials.Username
		botInfo.AccInfo.AccessToken = credentials.AccessToken
		id, err := servers.MakeFragBotServer(botInfo.BotId)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error occured while starting server"})
			return
		}

		botInfo.ServerId = id

		_, err = database.InsertDocument("accounts", botInfo)

		if err != nil {
			logging.LogWarn("Failed to add account credentials, error: " + err.Error())
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error:": "Failed to add credentials to database"})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"success": true})
	}
}
