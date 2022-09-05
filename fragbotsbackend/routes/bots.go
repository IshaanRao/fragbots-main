package routes

import (
	"errors"
	"fragbotsbackend/constants"
	"fragbotsbackend/database"
	"fragbotsbackend/fragaws"
	"fragbotsbackend/logging"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strconv"
)

// Bot Info Types

type BotInfo struct {
	BotId    string       `json:"botId" bson:"botId"`
	BotType  Bot          `json:"botType" bson:"botType"`
	ServerId string       `json:"serverId,omitempty" bson:"serverId,omitempty"`
	AccInfo  *AccountInfo `json:"accountInfo,omitempty" bson:"accInfo"`
	DiscInfo *DiscordInfo `json:"discordInfo" bson:"discInfo"`
}

type DiscordInfo struct {
	LogWebhook     string `json:"logWebhook"`
	ConsoleWebhook string `json:"consoleWebhook"`
}

type AccountInfo struct {
	UUID        string `json:"uuid" bson:"uuid"`
	Username    string `json:"username" bson:"username"`
	Email       string `json:"email" bson:"email"`
	Password    string `json:"password" bson:"password"`
	AccessToken string `json:"accessToken" bson:"accessToken"`
	AuthData    MSauth `json:"msAuth" bson:"authData"`
}

// Requests

type PostBotRequest struct {
	Stage    int    `json:"stage" form:"stage"`
	Email    string `json:"email,omitempty" form:"email"`
	Password string `json:"password,omitempty" form:"password"`
	UserCode string `json:"userCode,omitempty" form:"userCode"`
}

type CreateBotRequest struct {
	UserCode string `json:"userCode"`
}

type DeleteBotRequest struct {
	Delete bool `json:"delete" form:"delete"`
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
	Channel chan *MSauth
	BotData *BotInfo
}

type Bot string

const (
	Priority    Bot = "PRIORITY"
	Exclusive       = "EXCLUSIVE"
	Active          = "ACTIVE"
	Whitelisted     = "WHITELISTED"
	Verified        = "VERIFIED"
)

var authChannels = make(map[string]MsAuthChannelData)

func getVerifiedOneBotInfo() *BotInfo {
	return &BotInfo{
		BotId:   "Verified1",
		BotType: Verified,
		DiscInfo: &DiscordInfo{
			LogWebhook:     constants.VerifiedOneLogWebhook,
			ConsoleWebhook: constants.VerifiedOneConsoleWebhook,
		},
		AccInfo: &AccountInfo{},
	}
}

func getVerifiedTwoBotInfo() *BotInfo {
	return &BotInfo{
		BotId:   "Verified2",
		BotType: Verified,
		DiscInfo: &DiscordInfo{
			LogWebhook:     constants.VerifiedTwoLogWebhook,
			ConsoleWebhook: constants.VerifiedTwoConsoleWebhook,
		},
		AccInfo: &AccountInfo{},
	}
}

func getWhitelistedBotInfo() *BotInfo {
	return &BotInfo{
		BotId:   "Whitelisted",
		BotType: Whitelisted,
		DiscInfo: &DiscordInfo{
			LogWebhook:     constants.WhitelistedLogWebhook,
			ConsoleWebhook: constants.WhitelistedConsoleWebhook,
		},
		AccInfo: &AccountInfo{},
	}
}

func getActiveBotInfo() *BotInfo {
	return &BotInfo{
		BotId:   "Active",
		BotType: Active,
		DiscInfo: &DiscordInfo{
			LogWebhook:     constants.ActiveLogWebhook,
			ConsoleWebhook: constants.ActiveConsoleWebhook,
		},
		AccInfo: &AccountInfo{},
	}
}

func getExclusiveBotInfo() *BotInfo {
	return &BotInfo{
		BotId:   "Exclusive",
		BotType: Exclusive,
		DiscInfo: &DiscordInfo{
			LogWebhook:     constants.ExclusiveLogWebhook,
			ConsoleWebhook: constants.ExclusiveConsoleWebhook,
		},
		AccInfo: &AccountInfo{},
	}
}

func getPriorityBotInfo() *BotInfo {
	return &BotInfo{
		BotId:   "Priority",
		BotType: Priority,
		DiscInfo: &DiscordInfo{
			LogWebhook:     constants.PriorityLogWebhook,
			ConsoleWebhook: constants.PriorityConsoleWebhook,
		},
		AccInfo: &AccountInfo{},
	}
}

func getBotInfo(botId string) (*BotInfo, error) {
	var botInfo *BotInfo = nil
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

func GetBot(c *gin.Context) {
	id := c.Param("botid")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Missing BotID"})
		return
	}

	var botInfo BotInfo
	err := database.GetDocument("accounts", bson.D{{"botId", id}}, &botInfo)
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

func DeleteBot(c *gin.Context) {
	botId := c.Param("botid")
	var request DeleteBotRequest
	err := c.Bind(&request)
	if botId == "" || err != nil {
		logging.LogWarn("Stop bot request failed invalid request body error:" + err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	logging.Log("Got delete bot request, deleting: " + strconv.FormatBool(request.Delete))

	var botInfo BotInfo
	err = database.GetDocument("accounts", bson.D{{"botId", botId}}, &botInfo)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid BotID"})
		return
	}
	if botInfo.ServerId == "" {
		if request.Delete {
			err := database.UpdateDocument("accounts", bson.D{{"botId", botId}}, bson.D{{"botId", "archive_" + botId}})
			if err != nil {
				logging.LogWarn("Failed to archive document err, " + err.Error())
				c.IndentedJSON(http.StatusOK, gin.H{"error": "Failed to update document when archiving"})
				return
			}
			c.IndentedJSON(http.StatusOK, gin.H{"success": true})
			return
		}
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "No active bot"})
		return
	}

	err = fragaws.DeleteInstance(botInfo.ServerId)
	if err != nil {
		logging.LogWarn(err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Failed to delete instance"})
		return
	}
	if request.Delete {
		err := database.UpdateDocument("accounts", bson.D{{"botId", botId}}, bson.D{{"botId", "archive_" + botId}})
		if err != nil {
			logging.LogWarn("Failed to archive document err, " + err.Error())
			c.IndentedJSON(http.StatusOK, gin.H{"error": "Failed to update document when archiving"})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"success": true})
		return
	}
	err = database.UpdateDocumentDelField("accounts", bson.D{{"botId", botId}}, bson.D{{"serverId", nil}})
	if err != nil {
		c.IndentedJSON(http.StatusOK, gin.H{"error": "Failed to update document"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"success": true})

}

func PutBot(c *gin.Context) {
	botId := c.Param("botid")
	if botId == "" {
		logging.LogWarn("Put bot request failed invalid request body")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	var botInfo BotInfo
	err := database.GetDocument("accounts", bson.D{{"botId", botId}}, &botInfo)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid BotID"})
		return
	}
	if botInfo.ServerId != "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Bot currently active"})
		return
	}
	id, err := fragaws.MakeFragBotServer(botInfo.BotId)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "error occurred while starting server"})
		return
	}
	err = database.UpdateDocument("accounts", bson.D{{"botId", botId}}, bson.D{{"serverId", id}})
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "error occurred while updating serverid in db PLEASE TELL PRINCE"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"success": true})
}

func PostBot(c *gin.Context) {
	botId := c.Param("botid")
	var request PostBotRequest
	err := c.Bind(&request)

	if botId == "" || err != nil || request.Stage == 0 {
		logging.LogWarn("Post bot request failed invalid request body")
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	info, err := getBotInfo(botId)
	if err != nil {
		logging.LogWarn("Post bot request failed: " + err.Error())
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
		id, err := fragaws.MakeFragBotServer(botInfo.BotId)
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
