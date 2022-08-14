package main

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
)

type FragBotData struct {
	BotInfo struct {
		BotId       string `json:"botId"`
		BotType     Bot    `json:"botType"`
		AccountInfo struct {
			Uuid        string `json:"uuid"`
			Username    string `json:"username"`
			Email       string `json:"email"`
			Password    string `json:"password"`
			AccessToken string `json:"accessToken"`
		} `json:"accountInfo"`
		DiscordInfo struct {
			LogWebhook     string `json:"logWebhook"`
			ConsoleWebhook string `json:"consoleWebhook"`
		} `json:"discordInfo"`
	} `json:"botInfo"`
}

type HypixelStatusResponse struct {
	Success bool   `json:"success"`
	Uuid    string `json:"uuid"`
	Session struct {
		Online   bool   `json:"online"`
		GameType string `json:"gameType"`
		Mode     string `json:"mode"`
		Map      string `json:"map"`
	} `json:"session"`
}

type HypixelStatusError struct {
	Success bool   `json:"success"`
	Cause   string `json:"cause"`
}

type MojangResponse struct {
	Name string `json:"name"`
	UUID string `json:"id"`
}

type FragBotsUserResponse struct {
	Status int          `json:"status"`
	User   FragBotsUser `json:"data"`
}

type FragBotsUser struct {
	Id          string `json:"_id"`
	TimesUsed   int    `json:"timesused"`
	Discord     string `json:"discord"`
	Blacklisted bool   `json:"blacklisted"`
	Whitelisted bool   `json:"whitelisted"`
	Exclusive   bool   `json:"exclusive"`
	Active      bool   `json:"active"`
}

type Bot string

const (
	Exclusive   Bot = "EXCLUSIVE"
	Active          = "ACTIVE"
	Whitelisted     = "WHITELISTED"
	Verified        = "VERIFIED"
)

var fragbotBackend = "https://api.fragbots.xyz/v2"

func getFragBotsUser(username string) (*FragBotsUser, error) {
	mojData, err := getMojangData(username)
	if err != nil {
		return nil, err
	}
	userUUID, err := uuid.Parse(mojData.UUID)
	if err != nil {
		return nil, err
	}

	resp, err := get(fragbotBackend+"/users/"+userUUID.String(), nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 400 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("server error when getting fragbots user")
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fragBotResp := &FragBotsUserResponse{}
	err = json.Unmarshal(b, fragBotResp)
	if err != nil {
		return nil, err
	}

	return &fragBotResp.User, nil
}

func deleteBot() error {
	_, err := ReqClient.R().
		SetHeader("access-token", AccessToken).
		SetBodyJsonString("{\"botId\":\"" + FragData.BotInfo.BotId + "\"}").
		Post(BackendUrl + "/botinfo/removebot")
	return err
}
func isOnline(uuid string) (bool, error) {
	dataResp := HypixelStatusResponse{}
	errResp := HypixelStatusError{}
	_, err := ReqClient.R().
		SetHeader("ApiKey", HypixelApiKey).
		SetResult(dataResp).
		SetError(errResp).
		Post("https://api.hypixel.net/status" + uuid)

	if err != nil {
		return false, err
	}
	if !dataResp.Success {
		return false, errors.New(errResp.Cause)
	}
	return dataResp.Session.Online, nil
}

func addUse(uuid string) error {
	payload := strings.NewReader("uuid=" + uuid)

	client := &http.Client{}
	req, err := http.NewRequest("POST", fragbotBackend+"/uses/adduse", payload)
	if err != nil {
		return err
	}

	req.Header.Add("authkey", AuthKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New("adduse failed")
	}
	return nil
}

// getMojangData function used to convert username to uuid
func getMojangData(username string) (*MojangResponse, error) {
	resp, err := get("https://api.mojang.com/users/profiles/minecraft/"+username, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("mojang api did not respond, either invalid username cloudflare block or rate limit")
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	mojData := &MojangResponse{}
	err = json.Unmarshal(b, mojData)
	if err != nil {
		return nil, err
	}
	return mojData, nil
}
