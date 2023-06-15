package logging

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
)

// Color coded to make logs easier to read
var colorReset = "\033[0m"
var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"

var conn *websocket.Conn

// InitializeLogger gives logger the connection, so it can send messages to FragLink
func InitializeLogger(c *websocket.Conn) {
	conn = c
}

func LogWarn(v ...any) {
	log.Println(colorYellow+"[WARN]", fmt.Sprintln(v...), colorReset)
	SendLogCommand("[WARN] " + fmt.Sprint(v...))
}

func Log(v ...any) {
	log.Println(colorGreen+"[INFO]", fmt.Sprintln(v...), colorReset)
	SendLogCommand("[INFO] " + fmt.Sprint(v...))
}

func LogFatal(v ...any) {
	log.Println(colorRed+"[FATAL]", fmt.Sprintln(v...), colorReset)
	SendLogCommand("[FATAL] " + fmt.Sprint(v...))
	os.Exit(1)
}

// SendLogCommand sends a log message to fraglink
// doesnt use api package to avoid cyclic dependency
func SendLogCommand(logMsg string) {
	if conn == nil {
		return
	}
	command := map[string]interface{}{
		"Name": "Log",
		"Data": logMsg,
	}
	marshal, _ := json.Marshal(command)
	err := conn.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		log.Println("Failed to send log", err)
	}
	return
}
