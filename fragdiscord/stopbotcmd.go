package main

import "github.com/bwmarrin/discordgo"

var StopBotCommand = &Command{
	Name: "stopbot",
	BaseCommand: &discordgo.ApplicationCommand{
		Name:                     "stopbot",
		Description:              "Stops a fragbot",
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
					{
						Name:  "Priority",
						Value: "Priority",
					},
				},
				Required: true,
			},
		},
	},
	Handler:  stopBotRun,
	RunAsync: true,
}

func stopBotRun(client *discordgo.Session, event *discordgo.InteractionCreate) {
	id := event.ApplicationCommandData().Options[0].StringValue()
	err := StopBot(id)
	if err != nil {
		NewMessageBuilder().addEmbed(NewEmbedBuilder().setTitle("Error Occured").setDesc(err.Err)).sendAsInteractionResponseMessage(client, event)
		return
	}
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().
			setTitle("Success").
			setDesc("Stopped Fragbot with id: `"+id+"`")).
		sendAsInteractionResponseMessage(client, event)
}
