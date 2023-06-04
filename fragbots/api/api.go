package api

import (
	"context"
	"encoding/json"
	"github.com/Prince/fragbots/logging"
	"github.com/gorilla/websocket"
	"net/http"
)

const addr = "localhost:2468"

var srv = &http.Server{Addr: addr}
var upgrader = websocket.Upgrader{}

// client is websocket conn
// initialized when websocket connects
var wsClient *websocket.Conn

// StartApi starts the server that the websocket runs on
// blocks until error
func StartApi() error {
	http.HandleFunc("/ws", wsHandler)
	logging.Log("Starting server at:", addr)
	return srv.ListenAndServe()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		logging.LogWarn("Error upgrading to ws connection:", err)
		return
	}
	logging.LogWarn("Started websocket connection")
	defer func() {
		_ = c.Close()
		logging.Log("WS Client shut down, shutting down server")
		err := srv.Shutdown(context.TODO())
		if err != nil {
			logging.LogWarn("Error while shutting down server, reason:", err)
			return
		}
	}()

	wsClient = c

	sendCommand(WsCommand{Name: "Connected"})
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logging.LogWarn("Websocket Closed, reason:", err)
			return
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
		handleCommand(cmd)
	}
}
