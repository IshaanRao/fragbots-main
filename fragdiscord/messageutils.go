package main

import (
	json "encoding/json"
	"github.com/bwmarrin/discordgo"
	"time"
)

type MessageBuilder struct {
	Message *discordgo.InteractionResponseData
}

type ButtonBuilder struct {
	Button *discordgo.Button
}

type EmbedBuilder struct {
	Embed *discordgo.MessageEmbed
}

func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		Message: &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{},
			Embeds:     []*discordgo.MessageEmbed{},
		},
	}
}

func NewButtonBuilder() *ButtonBuilder {
	return &ButtonBuilder{
		Button: &discordgo.Button{},
	}
}

func NewEmbedBuilder() *EmbedBuilder {
	return &EmbedBuilder{
		Embed: &discordgo.MessageEmbed{
			Type:      discordgo.EmbedTypeRich,
			Timestamp: time.Now().Format(time.RFC3339),
			Color:     DefaultEmbedColor,
			Footer: &discordgo.MessageEmbedFooter{
				IconURL: FooterIcon,
				Text:    FooterText,
			},
		},
	}
}

func (messageBuilder *MessageBuilder) setContent(content string) *MessageBuilder {
	messageBuilder.Message.Content = content
	return messageBuilder
}

func (messageBuilder *MessageBuilder) makeEphemeral() *MessageBuilder {
	messageBuilder.Message.Flags = discordgo.MessageFlagsEphemeral
	return messageBuilder
}

func (messageBuilder *MessageBuilder) addEmbed(embedBuilder *EmbedBuilder) *MessageBuilder {
	messageBuilder.Message.Embeds = append(messageBuilder.Message.Embeds, embedBuilder.Embed)
	return messageBuilder
}

func (messageBuilder *MessageBuilder) addButton(actionRowIndex int, button *ButtonBuilder) *MessageBuilder {
	if len(messageBuilder.Message.Components) == actionRowIndex {
		messageBuilder.Message.Components = append(messageBuilder.Message.Components, discordgo.ActionsRow{})
	}
	actionRowBytes, err := messageBuilder.Message.Components[actionRowIndex].MarshalJSON()
	if err != nil {
		LogWarn("Failed to add button with id: " + button.Button.CustomID + ", error: " + err.Error())
		return messageBuilder
	}

	actionRow := discordgo.ActionsRow{}
	err = json.Unmarshal(actionRowBytes, &actionRow)
	if err != nil {
		LogWarn("Failed to add button with id: " + button.Button.CustomID + ", error: " + err.Error())
		return messageBuilder
	}
	actionRow.Components = append(actionRow.Components, button.Button)
	messageBuilder.Message.Components[actionRowIndex] = actionRow
	return messageBuilder
}

func (buttonBuilder *ButtonBuilder) setCustomID(customID string) *ButtonBuilder {
	buttonBuilder.Button.CustomID = customID
	return buttonBuilder
}

func (buttonBuilder *ButtonBuilder) setLabel(label string) *ButtonBuilder {
	buttonBuilder.Button.Label = label
	return buttonBuilder
}

func (buttonBuilder *ButtonBuilder) setStyle(style discordgo.ButtonStyle) *ButtonBuilder {
	buttonBuilder.Button.Style = style
	return buttonBuilder
}

func (embedBuilder *EmbedBuilder) setTitle(title string) *EmbedBuilder {
	embedBuilder.Embed.Title = title
	return embedBuilder
}

func (embedBuilder *EmbedBuilder) setType(embedType discordgo.EmbedType) *EmbedBuilder {
	embedBuilder.Embed.Type = embedType
	return embedBuilder
}

func (embedBuilder *EmbedBuilder) setDesc(desc string) *EmbedBuilder {
	embedBuilder.Embed.Description = desc
	return embedBuilder
}

func (embedBuilder *EmbedBuilder) setUrl(url string) *EmbedBuilder {
	embedBuilder.Embed.URL = url
	return embedBuilder
}

func (messageBuilder *MessageBuilder) sendAsInteractionResponseMessage(client *discordgo.Session, event *discordgo.InteractionCreate) {
	err := client.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: messageBuilder.Message,
	})
	if err != nil {
		LogWarn("Failed to send interactionresponsemessage, error: " + err.Error())
	}
}

func (messageBuilder *MessageBuilder) sendAsInteractionResponseEdit(client *discordgo.Session, event *discordgo.InteractionCreate) {
	_, err := client.InteractionResponseEdit(event.Interaction, &discordgo.WebhookEdit{
		Content:    &messageBuilder.Message.Content,
		Components: &messageBuilder.Message.Components,
		Embeds:     &messageBuilder.Message.Embeds,
	})
	if err != nil {
		LogWarn("Failed to send interactionresponseedit, error: " + err.Error())
	}
}

func (messageBuilder *MessageBuilder) sendAsInteractionResponseUpdate(client *discordgo.Session, event *discordgo.InteractionCreate) {
	err := client.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: messageBuilder.Message,
	})
	if err != nil {
		LogWarn("Failed to send interactionresponse, error: " + err.Error())
	}
}
