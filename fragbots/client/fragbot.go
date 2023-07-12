package client

import (
	"errors"
	"github.com/Prince/fragbots/logging"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/world"
	"github.com/Tnze/go-mc/chat"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Regex for checking if chat message is a party invite
var invRegex = regexp.MustCompile("(\\w*\\b) has invited you to join .* party!")

// Determines how long client should stay in party for
const (
	priorityWaitTime    = 10
	exclusiveWaitTime   = 10
	activeWaitTime      = 5
	whitelistedWaitTime = 5
	verifiedWaitTime    = 5
)

type FragBot struct {
	client    *bot.Client
	botWorld  *world.World
	Queue     *CmdQueue
	waitTime  int
	sentJoin  bool
	requester *Requester
	data      BotData
}

// initBot sets up everything to run fragbot logic
func newFragBot(c *bot.Client, data BotData) *FragBot {
	fb := FragBot{
		client:    c,
		Queue:     newCmdQueue(),
		sentJoin:  false,
		requester: data.Requester,
		data:      data,
	}

	switch data.BotType {
	case Priority:
		fb.waitTime = priorityWaitTime
	case Exclusive:
		fb.waitTime = exclusiveWaitTime
	case Active:
		fb.waitTime = activeWaitTime
	case Whitelisted:
		fb.waitTime = whitelistedWaitTime
	case Verified:
		fb.waitTime = verifiedWaitTime

	}
	fb.Queue.start()
	logging.Log("Initialized FragBot with client type:", data.BotType)
	return &fb
}

// onStart gets called whenever FragBot joins a server
func (fb *FragBot) onStart() error {
	if fb.sentJoin {
		return nil
	}
	fb.sentJoin = true
	logging.SendEmbed(fb.data.DiscInfo.LogWebhook, fb.data.AccountInfo.Username, "FragBot has joined Hypixel!")
	return nil
}

// onDc called when fragbots disconnected
func (fb *FragBot) onDc(reason chat.Message) error {
	logging.LogWarn("Bot Kicked Reason:" + reason.String())
	if strings.Contains(reason.String(), "banned") {
		return errors.New("bot banned: " + reason.String())
	}

	return errors.New("bot kicked:" + reason.String())
}

// onChat gets called when client receives a chat message
func (fb *FragBot) onChat(c chat.Message, _ bool) error {
	msg := c.ClearString()
	logging.Log(msg)

	// If there are no matches then the msg sent isn't a party invite
	if !invRegex.MatchString(msg) {
		return nil
	}

	username := invRegex.FindStringSubmatch(msg)[1]
	fb.onParty(username)

	return nil
}

// onParty gets called whenever fb is partied
// handles all logic for whether bot should join party and when
func (fb *FragBot) onParty(ign string) {
	fragBotUser, err := fb.requester.getFragBotsUser(ign)
	if err != nil {
		logging.LogWarn("Something went wrong when retrieving data for user: " + ign + ", Error: " + err.Error())
		return
	}
	botType := fb.data.BotType

	// Checks whether user is supposed to be able to party bot
	if fragBotUser == nil || (botType == Priority && !fragBotUser.Priority) || fragBotUser.Discord == "unknown" || ((!fragBotUser.Priority && !fragBotUser.Exclusive) && ((botType == Exclusive) || (botType == Whitelisted && !fragBotUser.Whitelisted) || (botType == Active && !fragBotUser.Active))) {
		logging.Log("(No Access) Rejected party invite from: " + ign)
	}
	queueLen := fb.Queue.GetTotalQueuedTasks()
	if (queueLen >= 10 && (botType == Verified || botType == Whitelisted || botType == Active)) || (queueLen >= 5 && (botType == Exclusive || botType == Priority)) {
		logging.SendEmbed(fb.data.DiscInfo.LogWebhook, fb.data.AccountInfo.Username, "Rejected party invite from: "+ign+", queue full!")
		logging.Log("(Queue Full) Rejected party invite from: " + ign)
		return
	}

	queueLen++

	logging.SendEmbedThumbnail(fb.data.DiscInfo.LogWebhook, fb.data.AccountInfo.Username, ign+" just partied "+fb.data.AccountInfo.Username+"!\nQueue Position: "+strconv.FormatInt(int64(queueLen), 10)+"\nEstimated Time: `"+strconv.FormatInt(int64(((queueLen-1)*(fb.waitTime+1))+1), 10)+" seconds"+"`", "https://mc-heads.net/avatar/"+ign)

	logging.Log("Received party invite from: " + ign)
	if err := fb.requester.addUse(fragBotUser.Id); err != nil {
		logging.LogWarn("AddUse failed for: "+ign+", err:", err)
	}

	// Queues user in fragbot command Queue
	fb.QueueCommand(ign)
	logging.Log("Successfully queued command for user: " + ign)
}

// QueueCommand queues a cmd to custom queue to be ran after prev cmd if there
func (fb *FragBot) QueueCommand(ign string) {
	fb.Queue.addTask(func() {
		logging.Log("Started processing of: " + ign + "'s invite, Queue Length: " + strconv.Itoa(fb.Queue.GetTotalQueuedTasks()))
		time.Sleep(1 * time.Second)
		logging.Log("Accepting invite from: " + ign)
		err := sendMsg(fb.client, "/party accept "+ign)
		if err != nil {
			logging.LogWarn("Failed to accept party from: "+ign+", error:", err)
			return
		}
		time.Sleep(time.Duration(fb.waitTime) * time.Second)
		logging.Log("Leaving party of: " + ign)
		err = sendMsg(fb.client, "/party leave")
		if err != nil {
			logging.LogWarn("Failed to leave party of: "+ign+", error: ", err)
		}
		return
	})
}

// stop stops all fb processes
func (fb *FragBot) stop() {
	fb.Queue.stop()
}
