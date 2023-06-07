package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/imroc/req/v3"
	"strconv"
	"time"
)

type FragInitData struct {
	BotId      string `json:"botId"`
	BotType    string `json:"botType"`
	WebhookUrl string `json:"webhookUrl"`
	AccInfo    struct {
		UUID        string `json:"uuid"`
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		AccessToken string `json:"accessToken"`
	} `json:"accountInfo"`
	ApiInfo apiInfo `json:"apiInfo"`
}

type apiInfo struct {
	BackendUrl  string `json:"backendUrl"`
	AccessToken string `json:"accessToken"`
}

// GetFragbotResp = resp from GET /v2/bots/botid
type GetFragbotResp struct {
	BotInfo struct {
		BotID   string `json:"botId"`
		BotType string `json:"botType"`
		AccInfo struct {
			UUID        string `json:"uuid"`
			Username    string `json:"username"`
			Email       string `json:"email"`
			Password    string `json:"password"`
			AccessToken string `json:"accessToken"`
		} `json:"accountInfo"`
		DiscInfo struct {
			LogWebhook     string `json:"logwebhook"`
			ConsoleWebhook string `json:"consolewebhook"`
		} `json:"discordInfo"`
	} `json:"botInfo"`
}

type WsCommand struct {
	Name string      `json:"name,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

var commands = map[string]func(data interface{}) error{
	"Error": handleError,
}

var ReqClient = req.C().
	SetTimeout(20 * time.Second)

func sendCommand(command WsCommand) {
	marshal, _ := json.Marshal(command)
	err := wsClient.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		LogFatal("Failed to write message to client: " + err.Error())
	}
	return
}

func handleCommand(command WsCommand) error {
	Log("Processing command with name: " + command.Name)
	f, ok := commands[command.Name]
	if !ok {
		sendCommand(WsCommand{Name: "Error", Data: "Invalid Command"})
		return nil
	}
	return f(command.Data)
}

func handleError(data interface{}) error {
	errMsg := fmt.Sprint(data)
	LogWarn("Recieved Error: " + errMsg)
	return nil
}

// sendStartBotCommand sends the command that starts the fragbot
func sendStartBotCommand() error {
	Log("Successfully connected to server, sending start bot command")
	FragData, err := getFragData(BotId)
	if err != nil {
		return err
	}

	var resp = FragInitData{
		BotId:      FragData.BotInfo.BotID,
		BotType:    FragData.BotInfo.BotType,
		AccInfo:    FragData.BotInfo.AccInfo,
		WebhookUrl: FragData.BotInfo.DiscInfo.LogWebhook,
		ApiInfo: apiInfo{
			BackendUrl:  BackendUrl,
			AccessToken: AccessToken,
		},
	}
	Log("Sending starting data to fragbot")
	dataString, _ := json.MarshalIndent(resp, "", "  ")
	Log(string(dataString))
	sendCommand(WsCommand{Name: "StartBot", Data: resp})
	return nil
}

func getFragData(botId string) (*GetFragbotResp, error) {
	var FragData GetFragbotResp

	if res, err := ReqClient.R().SetHeader("access-token", AccessToken).SetSuccessResult(&FragData).Get(BackendUrl + "/bots/" + botId); err != nil || res.StatusCode != 200 {
		if err == nil {
			LogWarn("Failed to get fragbots data, status code:" + strconv.Itoa(res.StatusCode) + ", res: " + res.String())
			return nil, errors.New("failed to get fragbot data")
		}
		LogWarn("Failed to get FragBotData error:", err.Error())
		return nil, err
	}
	return &FragData, nil
}
