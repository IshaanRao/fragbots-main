package main

import (
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/net/packet"
	"github.com/disgoorg/disgo/discord"
	"strings"
	"time"
)

// McClient holds data for the client
type McClient struct {
	Email    string
	Password string
	Started  bool
	Data     *UserData
	Client   *bot.Client
	Player   *basic.Player
	ShutDown bool
}

// UserData holds necessary data to log on to hypixel
type UserData struct {
	Username string
	Uuid     string
	Ssid     string
}

var serverIp = "play.hypixel.net"

// startClient starts the client to log on to hypixel
func (client *McClient) startClient() error {
	userData, err := client.getUserData()
	if err != nil {
		return err
	}

	client.Data = userData
	client.setupBot()
	err = client.joinHypixel()
	if err != nil {
		return err

	}

	return nil
}

// setupBot sets the necessary values for the client
func (client *McClient) setupBot() {
	client.Client = bot.NewClient()

	client.Client.Auth = bot.Auth{
		Name: client.Data.Username,
		UUID: client.Data.Uuid,
		AsTk: client.Data.Ssid,
	}

	client.Player = basic.NewPlayer(client.Client, basic.DefaultSettings)
}

// joinHypixel Makes FragBot join hypixel
func (client *McClient) joinHypixel() error {
	err := client.Client.JoinServer(serverIp)
	if err != nil {
		if strings.Contains(err.Error(), "banned") {
			_, err = logWebhook.CreateMessage(discord.NewWebhookMessageCreateBuilder().
				SetEmbeds(discord.NewEmbedBuilder().
					SetTitle(FragData.BotInfo.AccountInfo.Username+" Logs").
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
			return err
		} else {
			botLog("Kicked from hypixel while joining err: " + err.Error())
			return err
		}
	}

	return nil
}

// chat Sends chat messages through minecraft client
func (client *McClient) chat(msg string) error {
	return client.Client.Conn.WritePacket(packet.Marshal(
		0x03,
		packet.String(msg),
	))
}

// getUserData Gets data required for login from microsoft
func (client *McClient) getUserData() (*UserData, error) {
	data := UserData{
		Username: FragData.BotInfo.AccountInfo.Username,
		Uuid:     FragData.BotInfo.AccountInfo.Uuid,
		Ssid:     FragData.BotInfo.AccountInfo.AccessToken,
	}
	return &data, nil
}
