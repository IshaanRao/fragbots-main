package main

import "github.com/bwmarrin/discordgo"

var CreateBotCommand = &Command{
	Name: "createbot",
	BaseCommand: &discordgo.ApplicationCommand{
		Name:                     "createbot",
		Description:              "Adds bot credentials to fragbots database",
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
		},
	},
	Handler:  startBotRun,
	RunAsync: true,
}

func startBotRun(client *discordgo.Session, event *discordgo.InteractionCreate) {
	id := event.ApplicationCommandData().Options[0].StringValue()
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
		sendAsInteractionResponseEdit(client, event)

}
