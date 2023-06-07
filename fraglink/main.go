package main

import (
	"github.com/joho/godotenv"
	"os"
)

var BackendUrl string
var AccessToken string
var HypixelApiKey string
var BotId string

func main() {
	loadEnv()
	err := startApi()
	if err != nil {
		LogFatal("Client Stopped:", err)
	}
}

// loadEnv loads all environment variables
func loadEnv() {
	//loads .env file values into env
	err := godotenv.Load()
	if err != nil {
		LogWarn("NO ENV FILE FOUND MAY CAUSE ERRORS")
	}

	BackendUrl = getEnv("BACKEND_URI")
	AccessToken = getEnv("ACCESS_TOKEN")
	HypixelApiKey = getEnv("HYPIXEL_API_KEY")
	BotId = getEnv("BOT_ID")
}

// getEnv gets value from environment
// stops program if value is missing
func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		LogFatal("No " + key + "found in env")
	}
	return val
}
