package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

type WsCommand struct {
	Name string      `json:"name"`
	Data interface{} `json:"data,omitempty"`
}

type FragInitData struct {
	BotData       FragBotData
	BackendUrl    string
	AccessToken   string
	HypixelApiKey string
	BotId         string
}

var commands = map[string]func(data interface{}){
	"StartBot": startBotCmd,
	"Error":    handleError,
}
var upgrader = websocket.Upgrader{}
var connected = true
var client *websocket.Conn

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		botLogWarn("Error upgrading to ws connection: " + err.Error())
		return
	}
	botLog("Started websocket connection")
	connected = true
	client = c
	defer func() {
		c.Close()
		connected = false
		client = nil
	}()

	sendCommand(WsCommand{Name: "Connected"})
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			botLogFatal("Websocket Closed, reason: " + err.Error())
			return
		}
		cmd := WsCommand{}
		err = json.Unmarshal(message, &cmd)
		if err != nil {
			botLogWarn("Error parsing ws message: " + err.Error())
			sendCommand(WsCommand{Name: "Error", Data: "Something went wrong"})
			continue
		}
		if cmd.Name == "" {
			botLogWarn("Invalid command recieved: " + string(message))
			sendCommand(WsCommand{Name: "Error", Data: "Invalid Command"})
			continue
		}
		handleCommand(cmd)
	}
}

func sendCommand(command WsCommand) {
	marshal, _ := json.Marshal(command)
	err := client.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		botLogFatal("Failed to write message to client: " + err.Error())
	}
	return
}

func handleCommand(command WsCommand) {
	botLog("Processing command with name: " + command.Name)
	f, ok := commands[command.Name]
	if !ok {
		sendCommand(WsCommand{Name: "Error", Data: "Invalid Command"})
		return
	}
	f(command.Data)

	return
}

func startBotCmd(startData interface{}) {
	data := startData.(FragInitData)
	FragData = &data.BotData
	BackendUrl = data.BackendUrl
	AccessToken = data.AccessToken
	HypixelApiKey = data.HypixelApiKey
	BotId = data.BotId
	startBot()

}

func handleError(data interface{}) {
	var errMsg string
	_ = fmt.Sprintf(errMsg, data)
	botLogWarn("Recieved Error: " + errMsg)
}
