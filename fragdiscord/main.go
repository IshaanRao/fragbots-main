package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/imroc/req/v3"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var client *discordgo.Session
var ReqClient = req.C().
	SetTimeout(60 * time.Second)

func init() {

	err := loadToken()
	if err != nil {
		LogFatal(err.Error())
		return
	}

	client, err = discordgo.New("Bot " + Token)
	if err != nil {
		LogFatal("Failed to create bot" + err.Error())
	}
}

func main() {
	if DebugMode {
		LogWarn("Client started in Debug Mode change value in constants for production")
	}
	err := preStart()
	if err != nil {
		LogFatal("Something went wrong in prelogin: " + err.Error())
	}

	// Anything after this function will only run if there's an error if the proccess ends
	err = startBot()
	if err != nil {
		Log("Error has occurred while starting or shutting down bot: " + err.Error())
		return
	}

	Log("Discord Bot has successfully shut down.")
}

func preStart() error {
	err := addHandlers()
	if err != nil {
		return err
	}

	Debug("Finished preLogin function")
	return nil
}

func postStart() error {
	AddCommandsGuild(client)

	Debug("Finished postStart function")
	return nil
}

func addHandlers() error {
	addLoginHandler()
	LoadCommands(client)

	Debug("Added all handlers")
	return nil
}

func addLoginHandler() {
	client.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		Log("Bot Logged in to account: " + r.User.Username + ", with userId: " + r.User.ID)
	})
}

func loadToken() error {

	return nil
}

func startBot() error {
	err := client.Open()
	if err != nil {
		return err
	}
	Log("Bot Started")

	err = postStart()
	if err != nil {
		return err
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	err = client.Close()
	if err != nil {
		return err
	}
	return nil
}
