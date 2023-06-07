package client

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/Prince/fragbots/logging"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/bot/msg"
	"github.com/Tnze/go-mc/chat/sign"
	"github.com/Tnze/go-mc/data/packetid"
	pk "github.com/Tnze/go-mc/net/packet"
	"strings"
	"time"
)

// BotData stores all needed information to run fragbot
type BotData struct {
	BotId       string `json:"botId"`
	BotType     Bot    `json:"botType"`
	WebhookUrl  string `json:"webhookUrl"`
	AccountInfo struct {
		Uuid        string `json:"uuid"`
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		AccessToken string `json:"accessToken"`
	} `json:"accountInfo"`
	ApiInfo struct {
		BackendUrl  string `json:"backendUrl"`
		AccessToken string `json:"accessToken"`
	} `json:"apiInfo"`
}

type Bot string

const (
	Priority    Bot = "PRIORITY"
	Exclusive       = "EXCLUSIVE"
	Active          = "ACTIVE"
	Whitelisted     = "WHITELISTED"
	Verified        = "VERIFIED"
)

const serverIP = "mc.hypixel.net" //current hypixel IP

var fragBot *FragBot

// StartClient uses credentials to set up and start the fragbot
func StartClient(data BotData) error {
	c := bot.NewClient()

	c.Auth = bot.Auth{
		Name: data.AccountInfo.Username,
		UUID: data.AccountInfo.Uuid,
		AsTk: data.AccountInfo.AccessToken,
	}

	for {
		err := joinHypixel(c, data)
		if strings.Contains(err.Error(), "kicked") || strings.Contains(err.Error(), "EOF") {
			sendEmbed(data, "FragBot kicked from hypixel! Reconnecting...")

			fragBot.stop()
			time.Sleep(5 * time.Second) //Give bot some time before attempting reconnect
			continue
		}
		if strings.Contains(err.Error(), "banned") {
			sendEmbed(data, "FragBot BANNED from hypixel!")
		}
		return err
	}
}

// joinHypixel joins the server
// blocks until client is disconnected
func joinHypixel(c *bot.Client, data BotData) error {
	fragBot = newFragBot(c, data)

	player := basic.NewPlayer(c, basic.DefaultSettings, basic.EventsListener{SystemMsg: fragBot.onChat, Disconnect: fragBot.onDc, GameStart: fragBot.onStart}) //Registers all of fragbots hooks
	msg.New(c, player, msg.EventsHandler{})

	logging.Log("Joining Hypixel")
	err := c.JoinServer(serverIP)

	if err != nil {
		return err
	}
	logging.Log("Successfully joined Hypixel, starting main loop")
	for {
		if err = c.HandleGame(); err == nil {
			return errors.New("handle game returned nil")
		}

		//Handle Game only returns when there is an error so this handles the error
		if err2 := new(bot.PacketHandlerError); errors.As(err, err2) {
			//If bot is kicked or banned returns error so
			if strings.Contains(err2.Error(), "kicked") || strings.Contains(err2.Error(), "banned") {
				return err
			}
			logging.LogWarn("PacketHandlerError: ", err2)
		} else {
			return err
		}
	}
}

// sendMsg is an easy way to send messages
// code from: https://github.com/Tnze/go-mc/blob/v1.19.2/bot/msg/chat.go
func sendMsg(c *bot.Client, msg string) error {
	if len(msg) > 256 {
		return errors.New("message length greater than 256")
	}

	var salt int64
	if err := binary.Read(rand.Reader, binary.BigEndian, &salt); err != nil {
		return err
	}

	err := c.Conn.WritePacket(pk.Marshal(
		packetid.ServerboundChat,
		pk.String(msg),
		pk.Long(time.Now().UnixMilli()),
		pk.Long(salt),
		pk.ByteArray{},
		pk.Boolean(false),
		pk.Array([]sign.HistoryMessage{}),
		pk.Option[sign.HistoryMessage, *sign.HistoryMessage]{
			Has: false,
		},
	))
	return err
}
