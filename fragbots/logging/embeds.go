package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const FooterIcon = "https://cdn.discordapp.com/emojis/823999418592264232.webp?size=240&quality=lossless"
const DefaultEmbedColor = 3388927
const Version = "3.0.3"

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

// SendEmbed is an easier way to send an embed
func SendEmbed(webhook string, username string, description string) {
	err := SendMessage(webhook, Message{
		Embeds: []Embed{
			{
				Title:       username + " Logs",
				Description: description,
				Color:       DefaultEmbedColor,
				Footer: Footer{
					Text:    "FragBots " + Version,
					IconUrl: FooterIcon,
				},
				Timestamp: time.Now(),
			},
		},
	})

	if err != nil {
		LogWarn("Error sending embed:", err)
	}
}

// SendEmbedThumbnail is an easier way to send an embed w/ a thumbnail
func SendEmbedThumbnail(webhook string, username string, description string, thumbnailUrl string) {
	err := SendMessage(webhook, Message{
		Embeds: []Embed{
			{
				Title:       username + " Logs",
				Description: description,
				Color:       DefaultEmbedColor,
				Footer: Footer{
					Text:    "FragBots " + Version,
					IconUrl: FooterIcon,
				},
				Timestamp: time.Now(),
				Thumbnail: Thumbnail{Url: thumbnailUrl},
			},
		},
	})

	if err != nil {
		LogWarn("Error sending embed:", err)
	}
}
