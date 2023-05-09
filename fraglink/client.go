package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/imroc/req/v3"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var conn *websocket.Conn

var ReqClient = req.C().
	SetTimeout(20 * time.Second)

type FragInitData struct {
	BotData       interface{}
	BackendUrl    string
	AccessToken   string
	HypixelApiKey string
	BotId         string
}

type WsCommand struct {
	Name string      `json:"name,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

var commands = map[string]func(data interface{}){
	"Error":     handleError,
	"Connected": connected,
}

func start() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	Log("Connecting to ws at: " + u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	conn = c
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				LogFatal("Websocket Closed, reason: " + err.Error())
				return
			}
			cmd := WsCommand{}
			err = json.Unmarshal(message, &cmd)
			if err != nil {
				LogWarn("Error parsing ws message: " + err.Error())
				sendCommand(WsCommand{Name: "Error", Data: "Something went wrong"})
				continue
			}
			if cmd.Name == "" {
				LogWarn("Invalid command recieved: " + string(message))
				sendCommand(WsCommand{Name: "Error", Data: "Invalid Command"})
				continue
			}
			handleCommand(cmd)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			LogWarn("interrupt, gracefully closing connection")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Error occured on Client"))
			if err != nil {
				LogWarn("write close:" + err.Error())
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func sendCommand(command WsCommand) {
	marshal, _ := json.Marshal(command)
	err := conn.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		LogFatal("Failed to write message to client: " + err.Error())
	}
	return
}

func handleCommand(command WsCommand) {
	Log("Processing command with name: " + command.Name)
	f, ok := commands[command.Name]
	if !ok {
		sendCommand(WsCommand{Name: "Error", Data: "Invalid Command"})
		return
	}
	f(command.Data)

	return
}

func handleError(data interface{}) {
	var errMsg string
	_ = fmt.Sprintf(errMsg, data)
	LogWarn("Recieved Error: " + errMsg)
}

func connected(data interface{}) {
	Log("Successfully connected to server, sending start bot command")
	var FragData = getFragData(BotId)
	var resp = FragInitData{
		BotData:       FragData,
		BackendUrl:    BackendUrl,
		AccessToken:   AccessToken,
		HypixelApiKey: HypixelApiKey,
		BotId:         BotId,
	}
	sendCommand(WsCommand{Name: "StartBot", Data: resp})

}

func getFragData(botId string) interface{} {
	var FragData interface{}
	if res, err := ReqClient.R().SetHeader("access-token", AccessToken).SetSuccessResult(&FragData).Get(BackendUrl + "/bots/" + botId); err != nil || res.StatusCode != 200 {
		if err == nil {
			LogFatal("Failed to get fragbots data, status code:" + strconv.Itoa(res.StatusCode) + ", res: " + res.String())
		}
		LogFatal("Failed to get FragBotData error: " + err.Error())
	}
	return FragData
}
