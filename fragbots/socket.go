package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
)

type WsCommand struct {
	Name *string     `json:"name,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type WsError struct {
	Error string `json:"eraror"`
}

var upgrader = websocket.Upgrader{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		botLogWarn("Error upgrading to ws connection: " + err.Error())
		return
	}
	botLog("Started websocket connection")
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			botLogWarn("Error reading ws message: " + err.Error())
			break
		}
		cmd := WsCommand{}
		err = json.Unmarshal(message, &cmd)
		if err != nil {
			botLogWarn("Error parsing ws message: " + err.Error())
			marshal, _ := json.Marshal(WsError{Error: "Something went wrong"})
			err = c.WriteMessage(mt, marshal)
			if err != nil {
				botLogWarn("Error writing message to client: " + err.Error())
				break
			}
			continue
		}
		if cmd.Name == nil {
			botLogWarn("Invalid command recieved: " + string(message))
			marshal, _ := json.Marshal(WsError{Error: "Invalid Command"})
			err = c.WriteMessage(mt, marshal)
			if err != nil {
				botLogWarn("Error writing message to client: " + err.Error())
				break
			}
			continue
		}
		handleCommand(cmd, c)
	}
}

func handleCommand(command WsCommand, c *websocket.Conn) *error {
	botLog("Processing command with name: " + *command.Name)
	err := c.WriteMessage(1, []byte("aaaa"))
	if err != nil {
		botLogWarn("Error writing message to client: " + err.Error())
	}
	return nil
}
