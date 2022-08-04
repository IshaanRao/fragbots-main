package routes

import (
	"errors"
	"fragbotsbackend/constants"
	"fragbotsbackend/database"
	"fragbotsbackend/logging"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

type BotInfo struct {
	BotId       string           `json:"botId"`
	BotType     Bot              `json:"botType"`
	AccInfo     *AccountInfo     `json:"accountInfo,omitempty"`
	AccDocument *AccountDocument `json:"accountDocument,omitempty"`
	DiscInfo    *DiscordInfo     `json:"discordInfo"`
}

type DiscordInfo struct {
	LogWebhook     string `json:"logWebhook"`
	ConsoleWebhook string `json:"consoleWebhook"`
}

type AccountInfo struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	AccessToken string `json:"accessToken"`
}

type AccountDataResponse struct {
	Token   string `json:"token"`
	Profile struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"profile"`
}

type CredentialsRequest struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type RemoveBotRequest struct {
	BotId string `json:"botId"`
}

type AccountDocument struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	UsedOn   string `json:"usedOn" bson:"usedOn"`
}

type Bot string

const (
	Exclusive   Bot = "EXCLUSIVE"
	Active          = "ACTIVE"
	Whitelisted     = "WHITELISTED"
	Verified        = "VERIFIED"
)

func getVerifiedOneBotInfo() *BotInfo {
	return &BotInfo{
		BotId:   "Verified1",
		BotType: Verified,
		DiscInfo: &DiscordInfo{
			LogWebhook:     constants.VerifiedOneLogWebhook,
			ConsoleWebhook: constants.VerifiedOneConsoleWebhook,
		},
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
	}
}

func getBotInfo(botId string) (*BotInfo, error) {
	var botInfo *BotInfo = nil
	switch botId {
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
		return nil, nil
	}

	accDoc, accInfo, err := GetAccount(botId)
	if err != nil {
		return nil, errors.New("no account")
	}

	accDoc.UsedOn = botId
	err = database.UpdateDocument("accounts", bson.D{{"username", accDoc.Username}}, bson.D{{"usedOn", botId}})
	if err != nil {
		return nil, err
	}
	botInfo.AccDocument = accDoc
	botInfo.AccInfo = accInfo
	return botInfo, nil
}

func createBotStage1(c *gin.Context) {
	id := c.Param("botid")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Error": "Missing BotID"})
		return
	}
	logging.Debug("Creating fragbot with id: " + id)
	botInfo, err := getBotInfo(id)
	if err != nil {
		if err.Error() == "no account" {
			c.IndentedJSON(http.StatusServiceUnavailable, gin.H{"error": "no accounts"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	if botInfo.AccInfo == nil {
		c.IndentedJSON(http.StatusPartialContent, gin.H{"botInfo": botInfo})
		return
	}
	setupBot(botInfo)
}

func createBotStage2(c *gin.Context) {

}

func setupBot(botInfo *BotInfo) {

}

func getBotData(c *gin.Context) {
	id := c.Param("botid")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Error": "Missing BotID"})
		return
	}
	botInfo, err := getBotInfo(id)

	if err != nil {
		if err.Error() == "no account" {
			c.IndentedJSON(http.StatusServiceUnavailable, gin.H{"Error": "No accounts left"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Error": "Failed to change database"})
		logging.LogWarn(err.Error())
		return
	}

	if botInfo == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Error": "Invalid BotID"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"botId": id, "botInfo": botInfo})
}

func postRemoveCredentials(c *gin.Context) {
	var body RemoveBotRequest
	err := c.Bind(&body)
	if err != nil || body.BotId == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Error": "Invalid request body"})
		return
	}
	logging.Debug("Removing credentials from bot: " + body.BotId)
	success := database.DeleteDocument("accounts", bson.D{{"usedOn", body.BotId}})
	if success {
		c.IndentedJSON(http.StatusOK, gin.H{"Success": "Removed credentials from db"})
		return
	}
	c.IndentedJSON(http.StatusBadRequest, gin.H{"Error": "Failed to remove credentials"})

}

func postAddCredentials(c *gin.Context) {
	var credentials CredentialsRequest
	err := c.Bind(&credentials)
	if err != nil || (credentials.Password == "" || credentials.Username == "") {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Error": "Invalid request body"})
		return
	}
	logging.Debug("Received add credentials request username: " + credentials.Username + ", password: " + credentials.Password)
	if (database.DocumentExists("accounts", bson.D{{"Username", credentials.Username}})) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Error": "Account exists"})
		return
	}
	_, err = database.InsertDocument("accounts", AccountDocument{
		Username: credentials.Username,
		Password: credentials.Password,
		UsedOn:   "none",
	})

	if err != nil {
		logging.LogWarn("Failed to add account credentials, error: " + err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Error:": "Failed to add credentials to database"})
		return
	}
	logging.Debug("Added account credentials to server")
	c.IndentedJSON(http.StatusOK, gin.H{"Success:": "Added credentials to server"})
}

func GetAccount(botId string) (*AccountDocument, *AccountInfo, error) {
	account := AccountDocument{}
	err := database.GetDocument("accounts", bson.D{
		{"usedOn", botId},
	}, &account)

	if err != nil {
		if err2 := database.GetDocument("accounts", bson.D{{"usedOn", "none"}}, &account); err2 != nil {
			return nil, nil, err2
		}
	}

	accDataResp := AccountDataResponse{}
	get, err := constants.ReqClient.R().
		SetHeader("access-token", constants.AccessToken).
		SetHeader("username", account.Username).
		SetHeader("password", account.Password).
		SetResult(&accDataResp).
		Get(constants.AccountsURL + "/getaccdata")
	if err != nil || get.StatusCode != 200 {
		if err != nil {
			logging.LogWarn("Failed to get account data error: " + err.Error())
		}
		return &account, nil, nil
	}
	accInfo := AccountInfo{
		Username:    accDataResp.Profile.Name,
		UUID:        accDataResp.Profile.Id,
		AccessToken: accDataResp.Token,
	}
	return &account, &accInfo, nil
}
