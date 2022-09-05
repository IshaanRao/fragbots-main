package constants

import (
	"fragbotsbackend/logging"
	"github.com/imroc/req/v3"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

var Port int

var AccessToken string
var AuthKey string
var HypixelApiKey string
var MongoURL string

var AccountsURL string
var BackendUrl string

var ReqClient = req.C().
	SetTimeout(20 * time.Second)

var PriorityLogWebhook string
var PriorityConsoleWebhook string

var ExclusiveLogWebhook string
var ExclusiveConsoleWebhook string

var ActiveLogWebhook string
var ActiveConsoleWebhook string

var WhitelistedLogWebhook string
var WhitelistedConsoleWebhook string

var VerifiedOneLogWebhook string
var VerifiedOneConsoleWebhook string

var VerifiedTwoLogWebhook string
var VerifiedTwoConsoleWebhook string

func init() {
	err := godotenv.Load()
	if err != nil {
		logging.LogFatal("No .env file found")
	}

	Port, err = strconv.Atoi(getEnv("PORT"))
	if err != nil {
		logging.LogFatal("Invalid PORT env variable")
	}

	AccessToken = getEnv("ACCESS_TOKEN")
	AuthKey = getEnv("AUTHKEY")

	MongoURL = getEnv("MONGODB_URI")

	AccountsURL = getEnv("ACCOUNTS_URI")
	BackendUrl = getEnv("BACKEND_URI")
	HypixelApiKey = getEnv("HYPIXEL_API_KEY")

	PriorityLogWebhook = getEnv("PRIHOOK")
	PriorityConsoleWebhook = getEnv("PRICONSHOOK")

	ExclusiveLogWebhook = getEnv("EXCHOOK")
	ExclusiveConsoleWebhook = getEnv("EXCCONSHOOK")

	ActiveLogWebhook = getEnv("ACTHOOK")
	ActiveConsoleWebhook = getEnv("ACTCONSHOOK")

	WhitelistedLogWebhook = getEnv("WHITELISTEDHOOK")
	WhitelistedConsoleWebhook = getEnv("WHITELISTEDCONSHOOK")

	VerifiedOneLogWebhook = getEnv("VERHOOK")
	VerifiedOneConsoleWebhook = getEnv("VERCONSHOOK")

	VerifiedTwoLogWebhook = getEnv("VER2HOOK")
	VerifiedTwoConsoleWebhook = getEnv("VER2CONSHOOK")
}

func LoadConsts() {
	logging.Debug("Successfully loaded all constants")
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		logging.LogFatal("No" + key + "found in env")
	}
	return val
}
