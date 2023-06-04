package api

import (
	"encoding/json"
	"fmt"
	"github.com/Prince/fragbots/client"
	"github.com/Prince/fragbots/logging"
	"github.com/gorilla/websocket"
)

type WsCommand struct {
	Name string      `json:"name"`
	Data interface{} `json:"data,omitempty"`
}

// commands is a map that stores all commands that can be processed
var commands = map[string]func(data interface{}){
	"StartBot": startBotCmd,
	"Error":    handleError,
}

// handleCommand takes command name and calls the command handler
func handleCommand(command WsCommand) {
	logging.LogWarn("Processing command with name:" + command.Name)
	f, ok := commands[command.Name]
	if !ok {
		sendCommand(WsCommand{Name: "Error", Data: "Invalid Command"})
		return
	}
	f(command.Data)
	return
}

// sendCommand sends a WsCommand to webhook connection
func sendCommand(command WsCommand) {
	marshal, _ := json.Marshal(command)
	err := wsClient.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		logging.LogWarn("Failed to write message to client, shutting down ws:", err)
		err := wsClient.Close()
		if err != nil {
			logging.LogWarn("Error closing client:", err)
			return
		}
	}
	return
}

// startBotCmd starts the fragbot and receives the bot data
func startBotCmd(rawData interface{}) {
	var data client.BotData
	err := mapToInterface(rawData, &data)
	if err != nil {
		sendCommand(WsCommand{Name: "Error", Data: "Failed to parse starting data"})
		logging.LogFatal("Failed to parse starting data (should never happen)")
	}
	err = client.StartClient(data)
	if err != nil {
		logging.LogWarn("Error running bot, shutting down ws:", err)
		err := wsClient.Close()
		if err != nil {
			logging.LogWarn("Error closing client:", err)
		}
		return
	}

}

func mapToInterface(mapData any, inter any) error {
	data, err := json.Marshal(mapData)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, inter)
	return err
}

func handleError(data interface{}) {
	var errMsg string
	_ = fmt.Sprintf(errMsg, data)
	logging.LogWarn("Websocket received error:", errMsg)
}
