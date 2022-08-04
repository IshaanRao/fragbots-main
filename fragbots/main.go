package main

import (
	"encoding/json"
	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
	"os"
	"time"
)

var Client McClient
var FragData FragBotData
var BackendUrl string
var AccessToken string
var AuthKey string
var ReqClient = req.C().
	SetTimeout(20 * time.Second)

func main() {
	err := godotenv.Load()
	if err != nil {
		botLog("NO ENV FILE FOUND MAY CAUSE ERRORS")
	}

	BackendUrl = getEnv("BACKEND_URI")
	AccessToken = getEnv("ACCESS_TOKEN")
	AuthKey = getEnv("AUTHKEY")

	getFragData(getEnv("BOT_ID"))

	fbDataJson, err := json.MarshalIndent(FragData, "", "  ")
	if err != nil {
		botLogFatal("Something went wrong when serializing data: " + err.Error())
	}

	botLog("Successfully retrieved all data for FragBot starting bot...")
	botLog("Starting Data:")

	initWebhooks()

	botLog(string(fbDataJson))

	for {
		Client = McClient{}
		startFragBot()
		botLog("Restarting fragbot")
		time.Sleep(5 * time.Second)
	}
}

func getFragData(botId string) {
	if res, err := ReqClient.R().SetHeader("access-token", AccessToken).SetResult(&FragData).Get(BackendUrl + "/botinfo/" + botId); err != nil || res.StatusCode != 200 {
		botLogFatal("Failed to get FragBotData")
	}
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		botLogFatal("No" + key + "found in env")
	}
	return val
}
