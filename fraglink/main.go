package main

import (
	"github.com/joho/godotenv"
	"os"
)

var addr = "localhost:2468"

var BackendUrl string
var AccessToken string
var HypixelApiKey string
var BotId string

func main() {
	loadEnv()
	start()
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		LogWarn("NO ENV FILE FOUND MAY CAUSE ERRORS")
	}
	BackendUrl = getEnv("BACKEND_URI")
	AccessToken = getEnv("ACCESS_TOKEN")
	HypixelApiKey = getEnv("HYPIXEL_API_KEY")
	BotId = getEnv("BOT_ID")
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		LogFatal("No " + key + "found in env")
	}
	return val
}
