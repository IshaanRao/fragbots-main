package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Prince/fragbots/logging"
	"io/ioutil"
	"net/http"
	"time"
)

const FooterIcon = "https://cdn.discordapp.com/emojis/823999418592264232.webp?size=240&quality=lossless"
const DefaultEmbedColor = 3388927

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

type Hook struct {
	Embeds []Embed `json:"embeds,omitempty"`
}

// SendMessage sends the message to discord webhook link
func SendMessage(url string, message Message) error {
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

// sendEmbed is an easier way to send an embed
func sendEmbed(data BotData, description string) {
	err := SendMessage(data.WebhookUrl, Message{
		Embeds: []Embed{
			{
				Title:       data.AccountInfo.Username + " Logs",
				Description: description,
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
		logging.LogWarn("Error sending embed:", err)
	}
}

// sendEmbedThumbnail is an easier way to send an embed w/ a thumbnail
func sendEmbedThumbnail(data BotData, description string, thumbnailUrl string) {
	err := SendMessage(data.WebhookUrl, Message{
		Embeds: []Embed{
			{
				Title:       data.AccountInfo.Username + " Logs",
				Description: description,
				Color:       DefaultEmbedColor,
				Footer: Footer{
					Text:    "FragBots V3",
					IconUrl: FooterIcon,
				},
				Timestamp: time.Now(),
				Thumbnail: Thumbnail{Url: thumbnailUrl},
			},
		},
	})

	if err != nil {
		logging.LogWarn("Error sending embed:", err)
	}
}
