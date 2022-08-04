package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

type Command struct {
	Name                 string
	BaseCommand          *discordgo.ApplicationCommand
	Handler              func(client *discordgo.Session, event *discordgo.InteractionCreate)
	RunAsync             bool
	HasComponentHandlers bool
	ComponentHandlers    []*ComponentHandler
}

type ComponentHandler struct {
	CustomID  string
	Handler   func(client *discordgo.Session, event *discordgo.InteractionCreate)
	RunAsync  bool
	AnyoneRun bool
}

var commands = make(map[string]*Command)
var componentHandlers = make(map[string]*ComponentHandler)

func registerCommand(command *Command) {
	commands[command.Name] = command
	if command.HasComponentHandlers {
		for _, compHandler := range command.ComponentHandlers {
			componentHandlers[compHandler.CustomID] = compHandler
		}
	}
	Debug("Registered Command: " + command.Name)
}

func registerCommands() {
	registerCommand(AddBotCommand)

	Debug("Registered all commands successfully!")
}

// LoadCommands Loads the commands runs before client started
func LoadCommands(client *discordgo.Session) {
	Debug("Registering commands...")
	registerCommands()
	client.AddHandler(func(c *discordgo.Session, event *discordgo.InteractionCreate) {
		switch event.Type {
		case discordgo.InteractionApplicationCommand:
			if command, success := commands[event.ApplicationCommandData().Name]; success {
				Debug("User: " + event.Member.User.Username + " used command: " + command.Name)
				if command.RunAsync {
					go command.Handler(c, event)
					return
				}
				command.Handler(c, event)
			}
		case discordgo.InteractionMessageComponent:
			buttonId := strings.Split(event.MessageComponentData().CustomID, "-")[0]
			userId := strings.Split(event.MessageComponentData().CustomID, "-")[1]
			if compHandler, success := componentHandlers[buttonId]; success {
				Debug("User: " + event.Member.User.Username + " used component: " + compHandler.CustomID)
				if !compHandler.AnyoneRun && userId != event.Member.User.ID {
					NewMessageBuilder().
						setContent("This button isn't for you!").
						makeEphemeral().
						sendAsInteractionResponseMessage(client, event)
					return
				}
				if compHandler.RunAsync {
					go compHandler.Handler(c, event)
					return
				}
				compHandler.Handler(c, event)
			}
		}
	})
}

// AddCommandsGuild Adds slash commands runs after client starts
func AddCommandsGuild(client *discordgo.Session) {
	for name, cmd := range commands {
		_, err := client.ApplicationCommandCreate(client.State.User.ID, GuildId, cmd.BaseCommand)
		if err != nil {
			LogFatal("Failed to add command: " + name + err.Error())
		}
		Debug("Added " + name + " slash command to server")
	}
}
