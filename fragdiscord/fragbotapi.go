package main

import (
	"errors"
)

type CreateBotResponse struct {
	Err        string        `json:"error"`
	MsAuthInfo *AuthUserData `json:"msAuthInfo,omitempty"`
}
type AuthUserData struct {
	UserCode        string `json:"userCode"`
	VerificationUrl string `json:"VerificationUrl"`
	Email           string `json:"email"`
	Password        string `json:"password"`
}

type CreateBotRequest struct {
	UserCode string `json:"userCode"`
}

func CreateBot2(userCode string) error {
	resp, err := ReqClient.R().
		SetHeader("access-token", AccessToken).
		SetBodyJsonMarshal(CreateBotRequest{userCode}).
		Post(BackendUrl + "/botinfo/createbot2")
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("something went wrong")
	}
	return nil
}

func addBot(username string, password string) bool {
	post, err := ReqClient.R().
		SetHeader("access-token", AccessToken).
		SetBodyJsonString("{\"username\":\"" + username + "\",\"password\":\"" + password + "\"}").
		Post(BackendUrl + "/botinfo/addbot")
	if err != nil || post.StatusCode != 200 {
		if err != nil {
			LogWarn("Failed to add bot: " + err.Error())
		}
		return false
	}
	return true
}

func CreateBot(id string) *CreateBotResponse {
	result := CreateBotResponse{}
	_, err := ReqClient.R().
		SetHeader("access-token", AccessToken).
		SetResult(&result).
		SetError(&result).
		Post(BackendUrl + "/botinfo/createbot/" + id)
	if err != nil {
		return nil
	}
	return &result
}
