package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const FooterIcon = "https://cdn.discordapp.com/emojis/823999418592264232.webp?size=240&quality=lossless"
const DefaultEmbedColor = 3388927

var webhookLogQueue []string

// https://github.com/bensch777/discord-webhook-golang/blob/v0.0.5/discordwebhook.go

type Footer struct {
	Text    string `json:"text,omitempty"`
	IconUrl string `json:"icon_url"`
}
type Embed struct {
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Color       int       `json:"color,omitempty"`
	Footer      Footer    `json:"footer,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Thumbnail   Thumbnail `json:"thumbnail,omitempty"`
}

type Message struct {
	Embeds []Embed `json:"embeds,omitempty"`
}
type Thumbnail struct {
	Url string `json:"url"`
}

// sendMessage sends the message to discord webhook link
func sendMessage(url string, message Message) error {
	payload := new(bytes.Buffer)

	err := json.NewEncoder(payload).Encode(message)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", payload)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		defer resp.Body.Close()

		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf(string(responseBody))
	}

	return nil
}

// startWebhookLogger made to not hit 30 msg/minute limit for webhooks
// makes 24 reqs per min (60/2.5) to have wiggle room
func startWebhookLogger(webhookUrl string) {
	for {
		if len(webhookLogQueue) == 0 {
			continue
		}
		messages := webhookLogQueue
		webhookLogQueue = nil
		message := "```scss\n" + strings.Join(messages[:], "\n") + "\n```"
		err := sendMessage(webhookUrl, Message{
			Embeds: []Embed{
				{
					Title:       BotId + " Console",
					Description: message,
					Color:       DefaultEmbedColor,
					Footer: Footer{
						Text:    "FragBots V3",
						IconUrl: FooterIcon,
					},
					Timestamp: time.Now(),
				},
			},
		})
		if err != nil {
			LogWarn("Error sending message to console")
		}
		time.Sleep(2500 * time.Millisecond)
	}
}
