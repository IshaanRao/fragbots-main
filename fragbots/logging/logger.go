package logging

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"strings"
	"time"
)

// Color coded to make logs easier to read
var colorReset = "\033[0m"
var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"

var webhookLogQueue []string
var conn *websocket.Conn

// InitializeConsole gives logger the connection, so it can send messages to FragLink
func InitializeConsole(consoleUrl string, botId string) {
	startWebhookLogger(consoleUrl, botId)
}

func LogWarn(v ...any) {
	log.Println(colorYellow+"[WARN]", fmt.Sprintln(v...), colorReset)
	queueLog("[WARN] " + fmt.Sprint(v...))
}

func Log(v ...any) {
	log.Println(colorGreen+"[INFO]", fmt.Sprintln(v...), colorReset)
	queueLog("[INFO] " + fmt.Sprint(v...))
}

func LogFatal(v ...any) {
	log.Println(colorRed+"[FATAL]", fmt.Sprintln(v...), colorReset)
	queueLog("[FATAL] " + fmt.Sprint(v...))
	os.Exit(1)
}

// queueLog appends logging message to webhookLogQueue to be sent to console channel
func queueLog(logMsg string) {
	webhookLogQueue = append(webhookLogQueue, logMsg)
	return
}

// startWebhookLogger made to not hit 30 msg/minute limit for webhooks
// makes 24 reqs per min (60/2.5) to have wiggle room
func startWebhookLogger(webhookUrl string, botId string) {
	for {
		if len(webhookLogQueue) == 0 {
			continue
		}
		messages := webhookLogQueue
		webhookLogQueue = nil
		message := "```scss\n" + strings.Join(messages[:], "\n") + "\n```"
		err := SendMessage(webhookUrl, Message{
			Embeds: []Embed{
				{
					Title:       botId + " Console",
					Description: message,
					Color:       DefaultEmbedColor,
					Footer: Footer{
						Text:    "FragBots V3",
						IconUrl: FooterIcon,
					},
					Timestamp: time.Now(),
				},
			},
		})
		if err != nil {
			LogWarn("Error sending message to console:", err)
		}
		time.Sleep(2500 * time.Millisecond)
	}
}
