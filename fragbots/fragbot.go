package main

import (
	"github.com/Tnze/go-mc/chat"
	"github.com/disgoorg/disgo/discord"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Regex for checking if chat message is a party invite
var partyInviteRegex = regexp.MustCompile("(\\w*\\b) has invited you to join .* party!")

// Queue for handling fragbot party commands
var commandQueue FragQueue

// Wait time constants
var priorityWaitTime = 10
var exclusiveWaitTime = 10
var activeWaitTime = 5
var whitelistedWaitTime = 5
var verifiedWaitTime = 5
var waitTime = 0

var sentJoin bool

// startFragBot starts main fragbot program and creates the mc client
func startFragBot() {
	botLog("Reset variables and created command queue")
	commandQueue = NewFragQueue()
	sentJoin = false

	switch FragData.BotInfo.BotType {
	case Priority:
		waitTime = priorityWaitTime
	case Exclusive:
		waitTime = exclusiveWaitTime
	case Active:
		waitTime = activeWaitTime
	case Whitelisted:
		waitTime = whitelistedWaitTime
	case Verified:
		waitTime = verifiedWaitTime

	}
	commandQueue.Start()
	botLog("Starting client with wait time: " + strconv.FormatInt(int64(waitTime), 10))
	err := Client.startClient()
	if err != nil {
		botLog("error while starting client")
		botLog(err.Error())
		return
	}
}

// onChat function that gets called when bot recieves a chat message also calls fragbotparty function
func onChat(c chat.Message, _ bool) error {
	msg := c.ClearString()
	botLog(msg)

	if !partyInviteRegex.MatchString(msg) {
		return nil
	}
	onParty(partyInviteRegex.FindStringSubmatch(msg)[1])

	return nil
}

func stopBot() {
	err := Client.Client.Close()
	if err != nil {
		botLog("Failed to close client error: " + err.Error())
	}
	commandQueue.Stop()
}

func onStart() error {
	if sentJoin {
		return nil
	}
	sentJoin = true
	_, err := logWebhook.CreateMessage(discord.NewWebhookMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetTitle(Client.Client.Name+" Logs").
			SetDescription("FragBot has successfully logged on to Hypixel!").
			SetTimestamp(time.Now()).
			SetColor(DefaultEmbedColor).
			SetFooter("discord.gg/fragbots", FooterIcon).
			Build()).
		Build())
	if err != nil {
		botLog("Failed to send bot join webhook")
	}
	return nil
}

// onDc called when fragbots disconnected
func onDc(reason chat.Message) error {
	botLog("BOT KICKED REASON: " + reason.String())
	if strings.Contains(reason.String(), "banned") {
		_, err := logWebhook.CreateMessage(discord.NewWebhookMessageCreateBuilder().
			SetEmbeds(discord.NewEmbedBuilder().
				SetTitle(Client.Client.Name+" Logs").
				SetDescription("FragBot was banned from Hypixel!").
				SetTimestamp(time.Now()).
				SetColor(DefaultEmbedColor).
				SetFooter("discord.gg/fragbots", FooterIcon).
				Build()).
			Build())
		if err != nil {
			botLog("Failed to send fragbot banned webhook")
		}
		err = deleteBot()
		if err != nil {
			botLog("FAILED TO REMOVE BOT FROM DB")
		}
		botLogFatal("Bot was Banned")
	}
	stopBot()
	_, err := logWebhook.CreateMessage(discord.NewWebhookMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetTitle(Client.Client.Name+" Logs").
			SetDescription("FragBot was kicked from Hypixel! Reconnecting...").
			SetTimestamp(time.Now()).
			SetColor(DefaultEmbedColor).
			SetFooter("discord.gg/fragbots", FooterIcon).
			Build()).
		Build())
	if err != nil {
		botLog("Failed to send fragbot kicked webhook")
	}
	return nil
}

// onParty called when fragbot is partied does all handling of the invite
func onParty(ign string) {
	botName := Client.Client.Name
	fragBotUser, err := getFragBotsUser(ign)
	if err != nil {
		botLog("Something went wrong when retrieving data for user: " + ign + ", Error: " + err.Error())
		return
	}
	botType := FragData.BotInfo.BotType
	if fragBotUser == nil || (botType == Priority && !fragBotUser.Priority) || fragBotUser.Discord == "unknown" || ((!fragBotUser.Priority && !fragBotUser.Exclusive) && ((botType == Exclusive) || (botType == Whitelisted && !fragBotUser.Whitelisted) || (botType == Active && !fragBotUser.Active))) {
		botLog("(No Access) Rejected party invite from: " + ign)
		return
	}
	queueLen := commandQueue.GetTotalQueuedTasks()
	if (queueLen >= 10 && (botType == Verified || botType == Whitelisted || botType == Active)) || (queueLen >= 5 && (botType == Exclusive || botType == Priority)) {
		_, err = logWebhook.CreateMessage(discord.NewWebhookMessageCreateBuilder().
			SetEmbeds(discord.NewEmbedBuilder().
				SetTitle(botName+" Logs").
				SetDescription("Rejected party invite from: "+ign+", queue full!").
				SetTimestamp(time.Now()).
				SetThumbnail("https://mc-heads.net/avatar/"+ign).
				SetColor(DefaultEmbedColor).
				SetFooter("discord.gg/fragbots", FooterIcon).
				Build()).
			Build())
		if err != nil {
			botLog("Failed to send party rejected webhook")
		}
		botLog("(Queue Full) Rejected party invite from: " + ign)
		return
	}

	queueLen++

	_, err = logWebhook.CreateMessage(discord.NewWebhookMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetTitle(botName+" Logs").
			SetDescription(ign+" just partied "+botName+"!\nQueue Position: "+strconv.FormatInt(int64(queueLen), 10)+"\nEstimated Time: `"+strconv.FormatInt(int64(((queueLen-1)*(waitTime+1))+1), 10)+" seconds"+"`").
			SetThumbnail("https://mc-heads.net/avatar/"+ign).
			SetTimestamp(time.Now()).
			SetColor(DefaultEmbedColor).
			SetFooter("discord.gg/fragbots", FooterIcon).
			Build()).
		Build())
	if err != nil {
		botLog("Failed to send party received webhook")
	}

	botLog("Received party invite from: " + ign)
	if addUse(fragBotUser.Id) != nil {
		botLog("AddUse failed for: " + ign)
	}

	// Queues user in fragbot command Queue
	queueCommand(ign)
	botLog("Successfully queued command for user: " + ign)
}

func queueCommand(ign string) {
	commandQueue.AddTask(func() {
		botLog("Started processing of: " + ign + "'s invite, Queue Length: " + strconv.Itoa(commandQueue.GetTotalQueuedTasks()))
		time.Sleep(1000 * time.Millisecond)
		botLog("Accepting invite from: " + ign)
		err := Client.chat("/party accept " + ign)
		if err != nil {
			botLog("Error occurred while accepting party invite from: " + ign)
			botLog("Error: " + err.Error())
			return
		}
		time.Sleep(time.Duration(waitTime) * time.Second)
		botLog("Leaving party of: " + ign)
		err = Client.chat("/party leave")
		if err != nil {
			botLog("Error occurred while leaving party of: " + ign)
			botLog("Error: " + err.Error())
		}
		return
	})
}
