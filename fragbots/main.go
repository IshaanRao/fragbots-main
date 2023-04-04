package main

import (
	"encoding/json"
	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
	"time"
)

var Client *McClient
var FragData *FragBotData
var BackendUrl string
var AccessToken string
var HypixelApiKey string
var AuthKey string
var BotId string
var ReqClient = req.C().
	SetTimeout(20 * time.Second)

func main() {
	botLog("Docker test")
	err := godotenv.Load()
	if err != nil {
		botLog("NO ENV FILE FOUND MAY CAUSE ERRORS")
	}

	BackendUrl = getEnv("BACKEND_URI")
	AccessToken = getEnv("ACCESS_TOKEN")
	AuthKey = getEnv("AUTHKEY")
	HypixelApiKey = getEnv("HYPIXEL_API_KEY")
	BotId = getEnv("BOT_ID")

	go startBot()
	for {
		time.Sleep(30 * time.Second)
		if FragData == nil {
			continue
		}
		online, err := isOnline(FragData.BotInfo.AccountInfo.Uuid)
		if err != nil {
			botLog("Failed to get if bot was online: " + err.Error())
			continue
		}
		botLog("(Routine Bot Check) Bot Online: " + strconv.FormatBool(online))
		if online {
			continue
		}
		Client.ShutDown = true
		FragData = nil
		botLog("Waiting before starting fragbot")
		time.Sleep(30 * time.Second)
		go startBot()
	}

}

func startBot() {
	botLog("Starting fragbots")
	getFragData(BotId)
	fbDataJson, err := json.MarshalIndent(FragData, "", "  ")
	if err != nil {
		botLogFatal("Something went wrong when serializing data: " + err.Error())
	}

	botLog("Successfully retrieved all data for FragBot starting bot...")
	botLog("Starting Data:")
	botLog(string(fbDataJson))
	if Client == nil {
		initWebhooks()
	}
	defer func() {
		if err := recover(); err != nil {
			var ok bool
			err2, ok := err.(error)
			if !ok {
				return
			}
			if strings.Contains(err2.Error(), "banned") {
				panic("Bot is banned")
			}
			botLog("Fragbot goroutine panicked with error: " + err2.Error())
		}
	}()
	Client = &McClient{}
	startFragBot()
}

func getFragData(botId string) {
	if res, err := ReqClient.R().SetHeader("access-token", AccessToken).SetResult(&FragData).Get(BackendUrl + "/bots/" + botId); err != nil || res.StatusCode != 200 {
		botLogFatal("Failed to get FragBotData error: " + err.Error())
	}
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		botLogFatal("No" + key + "found in env")
	}
	return val
}
