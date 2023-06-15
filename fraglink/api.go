package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const addr = ":2468"

var srv = &http.Server{Addr: addr}
var upgrader = websocket.Upgrader{}

// client is websocket conn
// initialized when websocket connects
var wsClient *websocket.Conn

// startApi starts the server that the websocket runs on
// blocks until error or program is stopped
func startApi() error {
	http.HandleFunc("/ws", wsHandler)
	Log("Starting server at:", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	LogWarn("Shutting down server...")
	if wsClient != nil {
		err := wsClient.Close()
		LogWarn("Error while closing ws client:", err)
	}
	return srv.Close()
}

// wsHandler upgrades connection to ws connection and
// manages the fragbots commands
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
	data, err := sendStartBotCommand()
	if err != nil {
		LogWarn("Failed to send start bot cmd closing connection:", err)
		return
	}

	go startWebhookLogger(data.BotInfo.DiscInfo.ConsoleWebhook)

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
