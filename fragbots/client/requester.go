package client

import (
	"encoding/json"
	"errors"
	"github.com/Prince/fragbots/logging"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// MojangResponse is mojangs response when converting name to uuid
type MojangResponse struct {
	Name string `json:"name"`
	UUID string `json:"id"`
}

// GetFragbotResp is fragbots response containing all data for the fragbot
type GetFragbotResp struct {
	BotInfo BotData `json:"botInfo"`
}

// FragBotsUser is the data type for all fb user data
type FragBotsUser struct {
	Id          string `json:"_id"`
	TimesUsed   int    `json:"timesused"`
	Discord     string `json:"discord"`
	Blacklisted bool   `json:"blacklisted"`
	Whitelisted bool   `json:"whitelisted"`
	Exclusive   bool   `json:"exclusive"`
	Active      bool   `json:"active"`
	Priority    bool   `json:"priority,omitempty"`
}

// Requester helps make requests to fragbot api
type Requester struct {
	backendUrl  string
	accessToken string
}

// httpClient is a client for making all requests
var httpClient = http.Client{}

func NewRequester(backendUrl string, accessToken string) *Requester {
	r := &Requester{
		backendUrl:  backendUrl,
		accessToken: accessToken,
	}
	return r
}

// get easy way to send get reqs
func get(url string, headers *http.Header) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		for key, vals := range *headers {
			for _, val := range vals {
				req.Header.Add(key, val)
			}

		}
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetFragBotsUser retrieves user data from backend
func (r *Requester) getFragBotsUser(username string) (*FragBotsUser, error) {
	mojData, err := getMojangData(username)
	if err != nil {
		return nil, err
	}
	userUUID, err := uuid.Parse(mojData.UUID)
	if err != nil {
		return nil, err
	}

	resp, err := get(r.backendUrl+"/users/"+userUUID.String(), nil)
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
	fragBotResp := FragBotsUser{}
	err = json.Unmarshal(b, &fragBotResp)
	if err != nil {
		return nil, err
	}

	return &fragBotResp, nil
}

// AddUse adds a use to the use tracker for fragbots, so we can get uses/min stats
func (r *Requester) addUse(uuid string) error {
	payload := strings.NewReader("uuid=" + uuid)

	client := &http.Client{}
	req, err := http.NewRequest("POST", r.backendUrl+"/uses/", payload)
	if err != nil {
		return err
	}

	req.Header.Add("access-token", r.accessToken)
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

// GetFragData retrieves bot information from the backend
func (r *Requester) GetFragData(botId string) (*GetFragbotResp, error) {
	FragData := &GetFragbotResp{}

	res, err := get(r.backendUrl+"/bots/"+botId, &http.Header{
		"access-token": {r.accessToken},
	})
	if err != nil || res.StatusCode != 200 {
		if err == nil {
			return nil, errors.New("failed to get fragbot data status code: " + strconv.Itoa(res.StatusCode))
		}
		logging.LogWarn("Failed to get FragBotData error:", err.Error())
		return nil, err
	}

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, FragData)
	if err != nil {
		return nil, err
	}

	return FragData, nil
}
