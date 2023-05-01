package main

import (
	"encoding/json"
	"github.com/imroc/req/v3"
	"net/http"
	"strings"
	"time"
)

var Client *McClient
var FragData *FragBotData
var BackendUrl string
var AccessToken string
var HypixelApiKey string

var addr = "localhost:2468"

var BotId string
var ReqClient = req.C().
	SetTimeout(20 * time.Second)

func main() {
	startWsServer()
	/*for {
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
	}*/

}

func startBot() {
	botLog("Starting fragbots")
	fbDataJson, err := json.MarshalIndent(FragData, "", "  ")
	if err != nil {
		botLogFatal("Something went wrong when serializing data: " + err.Error())
	}
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

/*func getFragData(botId string) {
	if res, err := ReqClient.R().SetHeader("access-token", AccessToken).SetSuccessResult(&FragData).Get(BackendUrl + "/bots/" + botId); err != nil || res.StatusCode != 200 {
		if err == nil {
			botLog("Failed to get fragbots data, status code:" + strconv.Itoa(res.StatusCode) + ", res: " + res.String())
		}
		botLogFatal("Failed to get FragBotData error: " + err.Error())
	}
}*/

func startWsServer() {
	http.HandleFunc("/ws", wsHandler)
	botLog("Started websocket server at: " + addr)
	botLogFatal(http.ListenAndServe(addr, nil).Error())
}
