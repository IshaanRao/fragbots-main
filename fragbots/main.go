package main

import (
	"github.com/Prince/fragbots/client"
	"github.com/Prince/fragbots/logging"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	go func() {
		http.ListenAndServe(":1234", nil)
	}()
	logging.Log("Loading data from env")
	backendUrl := getEnv("BACKEND_URI")
	accessToken := getEnv("ACCESS_TOKEN")
	botId := getEnv("BOT_ID")

	requester := client.NewRequester(backendUrl, accessToken)

	data, err := requester.GetFragData(botId)

	if err != nil {
		logging.LogFatal("Failed to get fragbot data:", err)
	}

	data.BotInfo.Requester = requester

	err = client.StartClient(data.BotInfo)
	if err != nil {
		logging.LogFatal("Client stopped:", err)
	}
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		logging.LogFatal("No " + key + "found in env")
	}
	return val
}
