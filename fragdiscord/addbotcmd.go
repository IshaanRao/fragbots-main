package main

import (
	"github.com/bwmarrin/discordgo"
	"regexp"
)

var AddBotCommand = &Command{
	Name: "addbot",
	BaseCommand: &discordgo.ApplicationCommand{
		Name:                     "addbot",
		Description:              "Adds bot credentials to fragbots database",
		DefaultMemberPermissions: &AdminPerms,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "email",
				Description: "Email of microsoft account",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "password",
				Description: "Password of microsoft account",
				Required:    true,
			},
		},
	},
	Handler:              addBotRun,
	RunAsync:             false,
	HasComponentHandlers: true,
	ComponentHandlers: []*ComponentHandler{
		{
			CustomID: "ab_yes",
			Handler:  yesButton,
			RunAsync: true,
		},
		{
			CustomID: "ab_no",
			Handler:  noButton,
			RunAsync: false,
		},
	},
}

var credentialsRegex = regexp.MustCompile("`(.*?)`")

func addBotRun(client *discordgo.Session, event *discordgo.InteractionCreate) {
	username := event.ApplicationCommandData().Options[0].StringValue()
	password := event.ApplicationCommandData().Options[1].StringValue()
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().setTitle("AddBot Confirmation").setDesc("Are these credentials correct?\n"+"Username: `"+username+"`\n"+"Password: `"+password+"`")).
		addButton(0, NewButtonBuilder().setCustomID("ab_yes-"+event.Member.User.ID).setLabel("Yes").setStyle(discordgo.SuccessButton)).
		addButton(0, NewButtonBuilder().setCustomID("ab_no-"+event.Member.User.ID).setLabel("No").setStyle(discordgo.DangerButton)).
		makeEphemeral().
		sendAsInteractionResponseMessage(client, event)
}

func yesButton(client *discordgo.Session, event *discordgo.InteractionCreate) {
	description := event.Message.Embeds[0].Description
	regexMatches := credentialsRegex.FindAllStringSubmatch(description, 2)
	username := regexMatches[0][1]
	password := regexMatches[1][1]
	Debug("Adding Account to FragBots Database, Username: " + username + ", Password: " + password)
	if addBot(username, password) {
		NewMessageBuilder().
			addEmbed(NewEmbedBuilder().setTitle("Added Bot").setDesc("Added credentials to FragBots database!")).
			sendAsInteractionResponseUpdate(client, event)
	} else {
		NewMessageBuilder().
			addEmbed(NewEmbedBuilder().setTitle("FAILED").setDesc("Error occured while trying to add credentials!")).
			sendAsInteractionResponseUpdate(client, event)
	}

}

func noButton(client *discordgo.Session, event *discordgo.InteractionCreate) {
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().setTitle("Cancelled").setDesc("Canceled adding credentials to FragBots database!")).
		sendAsInteractionResponseUpdate(client, event)
}
