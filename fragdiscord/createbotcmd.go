package main

import (
	"github.com/bwmarrin/discordgo"
	"regexp"
)

var CreateBotCommand = &Command{
	Name: "createbot",
	BaseCommand: &discordgo.ApplicationCommand{
		Name:                     "createbot",
		Description:              "Creates fragbot",
		DefaultMemberPermissions: &AdminPerms,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "botid",
				Description: "BotId of fragbot",
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Verified 1",
						Value: "Verified1",
					},
					{
						Name:  "Verified 2",
						Value: "Verified2",
					},
					{
						Name:  "Whitelisted",
						Value: "Whitelisted",
					},
					{
						Name:  "Active",
						Value: "Active",
					},
					{
						Name:  "Exclusive",
						Value: "Exclusive",
					},
				},
				Required: true,
			},
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
	Handler:              startBotRun,
	RunAsync:             true,
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

func startBotRun(client *discordgo.Session, event *discordgo.InteractionCreate) {
	id := event.ApplicationCommandData().Options[0].StringValue()
	username := event.ApplicationCommandData().Options[1].StringValue()
	password := event.ApplicationCommandData().Options[2].StringValue()
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().setTitle("AddBot Confirmation").setDesc("Are these credentials correct?\n"+"Bot Id: `"+id+"`\n"+"Username: `"+username+"`\n"+"Password: `"+password+"`")).
		addButton(0, NewButtonBuilder().setCustomID("ab_yes-"+event.Member.User.ID).setLabel("Yes").setStyle(discordgo.SuccessButton)).
		addButton(0, NewButtonBuilder().setCustomID("ab_no-"+event.Member.User.ID).setLabel("No").setStyle(discordgo.DangerButton)).
		makeEphemeral().
		sendAsInteractionResponseMessage(client, event)
	/*id := event.ApplicationCommandData().Options[0].StringValue()
	NewMessageBuilder().
		makeEphemeral().
		addEmbed(NewEmbedBuilder().
			setTitle("Starting Bot").
			setDesc("Waiting for response from server...")).
		sendAsInteractionResponseMessage(client, event)
	res := CreateBot(id)
	if res == nil {
		NewMessageBuilder().addEmbed(NewEmbedBuilder().setTitle("Backend Offline")).sendAsInteractionResponseEdit(client, event)
		return
	}
	if res.Err != "" {
		errorMessageEmbed := NewEmbedBuilder().setTitle("Error Occured")
		switch res.Err {
		case "no accounts":
			errorMessageEmbed.setDesc("No accounts left please add using /addbot")
			break
		case "something went wrong":
			errorMessageEmbed.setDesc("Unexpected error occurred while creating bot")
			break
		}
		NewMessageBuilder().addEmbed(errorMessageEmbed).sendAsInteractionResponseEdit(client, event)
		return
	}
	if res.MsAuthInfo != nil {
		NewMessageBuilder().
			addEmbed(NewEmbedBuilder().
				setTitle("Authenticate Account").
				setDesc("Please enter details in the link provided in under a minute\n"+
					"Code: `"+res.MsAuthInfo.UserCode+"`\n"+
					"URL: "+res.MsAuthInfo.VerificationUrl+"\n"+
					"Email: `"+res.MsAuthInfo.Email+"`\n"+
					"Password: `"+res.MsAuthInfo.Password+"`")).
			sendAsInteractionResponseEdit(client, event)
		err := CreateBot2(res.MsAuthInfo.UserCode)
		if err != nil {
			NewMessageBuilder().addEmbed(NewEmbedBuilder().setTitle("Error Occured").setDesc("Unexpected error occurred while creating bot")).sendAsInteractionResponseEdit(client, event)
			LogWarn(err.Error())
			return
		}
	}
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().
			setTitle("Success").
			setDesc("Started Fragbot with id: `"+id+"`")).
		sendAsInteractionResponseEdit(client, event)*/

}

func yesButton(client *discordgo.Session, event *discordgo.InteractionCreate) {
	description := event.Message.Embeds[0].Description
	regexMatches := credentialsRegex.FindAllStringSubmatch(description, 3)
	botId := regexMatches[0][1]
	username := regexMatches[1][1]
	password := regexMatches[2][1]
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().
			setTitle("Starting Bot").
			setDesc("Waiting for response from server...")).
		sendAsInteractionResponseUpdate(client, event)
	res, err := CreateBot(botId, PostBotRequest{Stage: 1, Email: username, Password: password})
	if err != nil {
		NewMessageBuilder().addEmbed(NewEmbedBuilder().setTitle("Error Occurred").setDesc(err.Err)).sendAsInteractionResponseEdit(client, event)
		return
	}
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().
			setTitle("Authenticate Account").
			setDesc("Please enter details in the link provided in under a minute\n"+
				"Code: `"+res.MsAuthInfo.UserCode+"`\n"+
				"URL: "+res.MsAuthInfo.VerificationUrl+"\n"+
				"Email: `"+res.MsAuthInfo.Email+"`\n"+
				"Password: `"+res.MsAuthInfo.Password+"`")).
		sendAsInteractionResponseEdit(client, event)
	err = CreateBot2(botId, PostBotRequest{Stage: 2, UserCode: res.MsAuthInfo.UserCode})
	if err != nil {
		NewMessageBuilder().addEmbed(NewEmbedBuilder().setTitle("Error Occured").setDesc(err.Err)).sendAsInteractionResponseEdit(client, event)
		return
	}
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().
			setTitle("Success").
			setDesc("Started Fragbot with id: `"+botId+"`")).
		sendAsInteractionResponseEdit(client, event)

}

func noButton(client *discordgo.Session, event *discordgo.InteractionCreate) {
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().setTitle("Cancelled").setDesc("Canceled creating fragbot")).
		sendAsInteractionResponseUpdate(client, event)
}
