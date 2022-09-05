package main

import "github.com/bwmarrin/discordgo"

var DeleteBotCommand = &Command{
	Name: "deletebot",
	BaseCommand: &discordgo.ApplicationCommand{
		Name:                     "deletebot",
		Description:              "Deletes a fragbot",
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
	Handler:  deleteBotRun,
	RunAsync: true,
}

func deleteBotRun(client *discordgo.Session, event *discordgo.InteractionCreate) {
	id := event.ApplicationCommandData().Options[0].StringValue()
	err := DeleteBot(id)
	if err != nil {
		NewMessageBuilder().addEmbed(NewEmbedBuilder().setTitle("Error Occured").setDesc(err.Err)).sendAsInteractionResponseMessage(client, event)
		return
	}
	NewMessageBuilder().
		addEmbed(NewEmbedBuilder().
			setTitle("Success").
			setDesc("Deleted Fragbot with id: `"+id+"`")).
		sendAsInteractionResponseMessage(client, event)
}
