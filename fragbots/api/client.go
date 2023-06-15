package api

import (
	"encoding/json"
	"github.com/Prince/fragbots/logging"
	"github.com/gorilla/websocket"
	"net/url"
)

var conn *websocket.Conn

// addr is the address the fragbots socket server runs on
var addr = "fraglink:2468"

// StartClient starts wsClient that connects to FragLink ws server
func StartClient() error {

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	logging.Log("Connecting to fragbot ws at: " + u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logging.LogWarn("Failed to connect to ws")
		return err
	}
	conn = c
	logging.InitializeLogger(c)

	defer func() {
		err := c.Close()
		if err != nil {
			logging.LogWarn("Error while closing client:", err.Error())
		}
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logging.LogWarn("Websocket Closed, reason:", err)
			return err
		}

		//Converts ws message to WsCommand
		cmd := WsCommand{}
		err = json.Unmarshal(message, &cmd)
		if err != nil {
			logging.LogWarn("Error parsing ws message:", err)
			sendCommand(WsCommand{Name: "Error", Data: "Something went wrong"})
			continue
		}
		if cmd.Name == "" {
			logging.LogWarn("Invalid command received:" + string(message))
			sendCommand(WsCommand{Name: "Error", Data: "Invalid Command"})
			continue
		}

		err = handleCommand(cmd)
		if err != nil {
			return err
		}
	}

}
