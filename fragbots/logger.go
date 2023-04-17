package main

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/snowflake/v2"
	"log"
	"regexp"
	"time"
)

const FooterIcon = "https://cdn.discordapp.com/emojis/823999418592264232.webp?size=240&quality=lossless"
const FooterText = "FragBots"
const DefaultEmbedColor = 15954943

var colorReset = "\033[0m"

var colorRed = "\033[31m"
var colorGreen = "\033[32m"

var colorYellow = "\033[33m"

var webhookRegex = regexp.MustCompile("hooks/(.*)/(.*)")

var logWebhook webhook.Client
var consoleWebhook webhook.Client

var consoleCache = make([]string, 0)
var consoleDumpsActive = true

var intialized = false

// botLog logs data to the discord console and console logs
func botLog(msg string) {
	var prefix string
	if Client == nil || Client.Data == nil {
		prefix = "[NotLoggedIn]"
	} else {
		prefix = "[" + Client.Data.Username + "]"
	}
	fullMsg := colorGreen + prefix + " " + msg + colorReset
	println(fullMsg)
	consoleCache = append(consoleCache, prefix+" "+msg)
}

func botLogWarn(msg string) {
	var prefix string
	if Client == nil || Client.Data == nil {
		prefix = "[NotLoggedIn-Warn]"
	} else {
		prefix = "[" + Client.Data.Username + "-Warn]"
	}
	fullMsg := colorYellow + prefix + " " + msg + colorReset
	println(fullMsg)
	consoleCache = append(consoleCache, prefix+" "+msg)
}
func botLogFatal(msg string) {
	var prefix string
	if Client == nil || Client.Data == nil {
		prefix = "[NotLoggedIn-Fatal]"
	} else {
		prefix = "[" + Client.Data.Username + "-Fatal]"
	}

	fullMsg := colorRed + prefix + " " + msg + colorReset
	consoleCache = append(consoleCache, prefix+" "+msg)
	if intialized {
		DumpCache()
	}
	log.Fatal(fullMsg)
}

func initWebhooks() {
	loggerMatches := webhookRegex.FindAllStringSubmatch(FragData.BotInfo.DiscordInfo.LogWebhook, 3)[0]
	consoleMatches := webhookRegex.FindAllStringSubmatch(FragData.BotInfo.DiscordInfo.ConsoleWebhook, 3)[0]
	logWebhook = webhook.New(snowflake.MustParse(loggerMatches[1]), loggerMatches[2])
	consoleWebhook = webhook.New(snowflake.MustParse(consoleMatches[1]), consoleMatches[2])
	intialized = true
	go func() {
		for consoleDumpsActive {
			DumpCache()
			time.Sleep(10 * time.Second)
		}
	}()
}

func DumpCache() {
	if len(consoleCache) == 0 {
		return
	}
	dump := "```scss\n"
	for _, s := range consoleCache {
		dump += s + "\n"
	}
	dump += "```"
	consoleCache = nil
	_, err := consoleWebhook.CreateMessage(discord.NewWebhookMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetTitle(BotId+" Logs").
			SetDescription(dump).
			SetColor(DefaultEmbedColor).
			SetTimestamp(time.Now()).
			SetFooter(FooterText, FooterIcon).
			Build()).
		Build())
	if err != nil {
		println("Failed to send webhooks message!!! " + err.Error())
	}
}
