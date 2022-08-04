package main

import (
	"context"
	"errors"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/chat"
	"github.com/disgoorg/disgo/discord"
	"github.com/golang-queue/queue"
	"github.com/google/uuid"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Regex for checking if chat message is a party invite
var partyInviteRegex = regexp.MustCompile("(\\w*\\b) has invited you to join .* party!")

// Queue for handling fragbot party commands
var commandQueue *queue.Queue

// Wait time constants
var exclusiveWaitTime = 11
var activeWaitTime = 7
var whitelistedWaitTime = 6
var verifiedWaitTime = 5
var waitTime = 0

var queueLen int

var sentJoin bool

// startFragBot starts main fragbot program and creates the mc client
func startFragBot() {

	commandQueue = queue.NewPool(1)
	sentJoin = false
	queueLen = 0

	switch FragData.BotInfo.BotType {
	case Exclusive:
		waitTime = exclusiveWaitTime
	case Active:
		waitTime = activeWaitTime
	case Whitelisted:
		waitTime = whitelistedWaitTime
	case Verified:
		waitTime = verifiedWaitTime

	}

	err := Client.startClient()
	if err != nil {
		println(err.Error())
		return
	}

	basic.EventsListener{ChatMsg: onChat, Disconnect: onDc, GameStart: onStart}.Attach(Client.Client)

	for {
		if err = Client.Client.HandleGame(); err == nil {
			panic("Unexpected error has occurred!!")
			return
		}

		if err2 := new(bot.PacketHandlerError); errors.As(err, err2) {
			if err := new(bot.DisconnectErr); errors.As(err2, err) {
				println("Disconnect: ", err.Error())
				return
			} else {
				botLog("Unexpected Error: " + err.Error())
				return
			}
		} else {
			botLog("Unexpected Error: " + err.Error())
			return
		}
	}
}

// onChat function that gets called when bot recieves a chat message also calls fragbotparty function
func onChat(c chat.Message, _ byte, _ uuid.UUID) error {
	msg := c.ClearString()
	botLog(msg)

	if !partyInviteRegex.MatchString(msg) {
		return nil
	}
	onParty(partyInviteRegex.FindStringSubmatch(msg)[1])

	return nil
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
	if fragBotUser == nil || (botType == Exclusive && !fragBotUser.Exclusive) || (botType == Active && !fragBotUser.Active) || (botType == Whitelisted && !fragBotUser.Whitelisted) || (botType == Verified && fragBotUser.Discord == "unknown") {
		botLog("(No Access) Rejected party invite from: " + ign)
		return
	}
	if queueLen >= 10 {
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
			SetDescription(ign+"just partied"+botName+"!\nQueue Position: "+strconv.FormatInt(int64(queueLen), 10)+"\nEstimated Time: `"+strconv.FormatInt(int64((queueLen-1)*5), 10)+" seconds"+"`").
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
	if queueCommand(ign) != nil {
		botLog("Failed to queue command for user: " + ign)
		return
	}

}

func queueCommand(ign string) error {
	err := commandQueue.QueueTask(func(ctx context.Context) error {
		time.Sleep(250 * time.Millisecond)
		err := Client.chat("/party accept " + ign)
		if err != nil {
			queueLen--
			botLog("Error occurred while accepting party invite from: " + ign)
			return nil
		}
		time.Sleep(time.Duration(waitTime) * time.Second)
		err = Client.chat("/party leave")
		if err != nil {
			botLog("Error occurred while accepting leaving party of: " + ign)
		}
		queueLen--
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
