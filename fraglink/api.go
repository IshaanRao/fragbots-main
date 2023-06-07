package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
)

const addr = ":2468"

var srv = &http.Server{Addr: addr}
var upgrader = websocket.Upgrader{}

// client is websocket conn
// initialized when websocket connects
var wsClient *websocket.Conn

// startApi starts the server that the websocket runs on
// blocks until error
func startApi() error {
	http.HandleFunc("/ws", wsHandler)
	Log("Starting server at:", addr)
	return srv.ListenAndServe()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		LogWarn("Error upgrading to ws connection:", err)
		return
	}
	Log("Started websocket connection")
	defer func() {
		wsClient = nil
		err = c.Close()
		if err != nil {
			LogWarn("Error while closing ws, reason:", err)
		}
		return
	}()

	wsClient = c

	//tell bot to start
	err = sendStartBotCommand()
	if err != nil {
		LogWarn("Failed to send start bot cmd closing connection:", err)
		return
	}

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			LogWarn("Websocket Closed, reason:", err)
			return
		}

		//Converts ws message to WsCommand
		cmd := WsCommand{}
		err = json.Unmarshal(message, &cmd)
		if err != nil {
			LogWarn("Error parsing ws message:", err)
			sendCommand(WsCommand{Name: "Error", Data: "Something went wrong"})
			continue
		}
		if cmd.Name == "" {
			LogWarn("Invalid command received:" + string(message))
			sendCommand(WsCommand{Name: "Error", Data: "Invalid Command"})
			continue
		}
		err = handleCommand(cmd)
		if err != nil {
			LogWarn("Error while processing cmd, closing ws:", err)
			return
		}
	}
}
