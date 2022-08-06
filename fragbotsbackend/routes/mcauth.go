package routes

// Go MC MicroSoft Auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// MSauth holds Microsoft auth credentials
type MSauth struct {
	AccessToken  string
	ExpiresAfter int64
	RefreshToken string
}

type AuthUserData struct {
	UserCode        string `json:"userCode"`
	VerificationUrl string `json:"VerificationUrl"`
	Email           string `json:"email"`
	Password        string `json:"password"`
}

// AzureClientIDEnvVar Used to lookup Azure client id via os.Getenv if cid is not passed
const AzureClientIDEnvVar = "AzureClientID"

// AuthMSdevice Attempts to authorize user via device flow. Will block thread until gets error, timeout or actual authorization
func AuthMSdevice() (*AuthUserData, chan *MSauth, error) {
	cid := "88650e7e-efee-4857-b9a9-cf580a00ef43"

	DeviceResp, err := http.PostForm("https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode", url.Values{
		"client_id": {cid},
		"scope":     {`XboxLive.signin offline_access`},
	})
	if err != nil {
		return nil, nil, err
	}
	var DeviceRes map[string]interface{}
	json.NewDecoder(DeviceResp.Body).Decode(&DeviceRes)
	DeviceResp.Body.Close()
	if DeviceResp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("MS device request answered not HTTP200! Instead got %s and following json: %#v", DeviceResp.Status, DeviceRes)
	}
	DeviceCode, ok := DeviceRes["device_code"].(string)
	if !ok {
		return nil, nil, errors.New("Device code not found in response")
	}
	UserCode, ok := DeviceRes["user_code"].(string)
	if !ok {
		return nil, nil, errors.New("User code not found in response")
	}
	VerificationURI, ok := DeviceRes["verification_uri"].(string)
	if !ok {
		return nil, nil, errors.New("Verification URI not found in response")
	}
	_, ok = DeviceRes["expires_in"].(float64)
	if !ok {
		return nil, nil, errors.New("Expires In not found in response")
	}
	PoolInterval, ok := DeviceRes["interval"].(float64)
	if !ok {
		return nil, nil, errors.New("Pooling interval not found in response")
	}

	msDataChan := make(chan *MSauth)

	go func() {
		time.Sleep(4 * time.Second)
		for {
			time.Sleep(time.Duration(int(PoolInterval)+1) * time.Second)
			CodeResp, err := http.PostForm("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", url.Values{
				"client_id":   {cid},
				"scope":       {"XboxLive.signin offline_access"},
				"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
				"device_code": {DeviceCode},
			})
			if err != nil {
				msDataChan <- nil
				return
			}
			var CodeRes map[string]interface{}
			json.NewDecoder(CodeResp.Body).Decode(&CodeRes)
			CodeResp.Body.Close()
			if CodeResp.StatusCode == 400 {
				PoolError, ok := CodeRes["error"].(string)
				if !ok {
					msDataChan <- nil
					return
				}
				if PoolError == "authorization_pending" {
					continue
				}
				if PoolError == "authorization_declined" {
					msDataChan <- nil
					return
				}
				if PoolError == "expired_token" {
					msDataChan <- nil
					return
				}
				if PoolError == "invalid_grant" {
					msDataChan <- nil
					return
				}
			} else if CodeResp.StatusCode == 200 {
				auth := MSauth{}
				MSaccessToken, ok := CodeRes["access_token"].(string)
				if !ok {
					msDataChan <- nil
					return
				}
				auth.AccessToken = MSaccessToken
				MSrefreshToken, ok := CodeRes["refresh_token"].(string)
				if !ok {
					msDataChan <- nil
					return
				}
				auth.RefreshToken = MSrefreshToken
				MSexpireSeconds, ok := CodeRes["expires_in"].(float64)
				if !ok {
					msDataChan <- nil
					return
				}
				auth.ExpiresAfter = time.Now().Unix() + int64(MSexpireSeconds)
				msDataChan <- &auth
			} else {
				msDataChan <- nil
				return
			}
		}

	}()

	return &AuthUserData{
		UserCode:        UserCode,
		VerificationUrl: VerificationURI,
	}, msDataChan, nil

}

// AuthXBL Gets XBox Live token from Microsoft token
func AuthXBL(MStoken string) (string, error) {
	XBLdataMap := map[string]interface{}{
		"Properties": map[string]interface{}{
			"AuthMethod": "RPS",
			"SiteName":   "user.auth.xboxlive.com",
			"RpsTicket":  "d=" + MStoken,
		},
		"RelyingParty": "http://auth.xboxlive.com",
		"TokenType":    "JWT",
	}
	XBLdata, err := json.Marshal(XBLdataMap)
	if err != nil {
		return "", err
	}
	XBLreq, err := http.NewRequest(http.MethodPost, "https://user.auth.xboxlive.com/user/authenticate", bytes.NewBuffer(XBLdata))
	if err != nil {
		return "", err
	}
	XBLreq.Header.Set("Content-Type", "application/json")
	XBLreq.Header.Set("Accept", "application/json")
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateOnceAsClient,
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
	XBLresp, err := client.Do(XBLreq)
	if err != nil {
		return "", err
	}
	var XBLres map[string]interface{}
	json.NewDecoder(XBLresp.Body).Decode(&XBLres)
	XBLresp.Body.Close()
	if XBLresp.StatusCode != 200 {
		return "", fmt.Errorf("XBL answered not HTTP200! Instead got %s and following json: %#v", XBLresp.Status, XBLres)
	}
	XBLtoken, ok := XBLres["Token"].(string)
	if !ok {
		return "", errors.New("Token not found in XBL response")
	}
	return XBLtoken, nil
}

// XSTSauth Holds XSTS token and UHS
type XSTSauth struct {
	Token string
	UHS   string
}

// AuthXSTS Gets XSTS token using XBL
func AuthXSTS(XBLtoken string) (XSTSauth, error) {
	var auth XSTSauth
	XSTSdataMap := map[string]interface{}{
		"Properties": map[string]interface{}{
			"SandboxId":  "RETAIL",
			"UserTokens": []string{XBLtoken},
		},
		"RelyingParty": "rp://api.minecraftservices.com/",
		"TokenType":    "JWT",
	}
	XSTSdata, err := json.Marshal(XSTSdataMap)
	if err != nil {
		return auth, err
	}
	XSTSreq, err := http.NewRequest(http.MethodPost, "https://xsts.auth.xboxlive.com/xsts/authorize", bytes.NewBuffer(XSTSdata))
	if err != nil {
		return auth, err
	}
	XSTSreq.Header.Set("Content-Type", "application/json")
	XSTSreq.Header.Set("Accept", "application/json")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	XSTSresp, err := client.Do(XSTSreq)
	if err != nil {
		return auth, err
	}
	var XSTSres map[string]interface{}
	json.NewDecoder(XSTSresp.Body).Decode(&XSTSres)
	XSTSresp.Body.Close()
	if XSTSresp.StatusCode != 200 {
		return auth, fmt.Errorf("XSTS answered not HTTP200! Instead got %s and following json: %#v", XSTSresp.Status, XSTSres)
	}
	XSTStoken, ok := XSTSres["Token"].(string)
	if !ok {
		return auth, errors.New("Could not find Token in XSTS response")
	}
	auth.Token = XSTStoken
	XSTSdc, ok := XSTSres["DisplayClaims"].(map[string]interface{})
	if !ok {
		return auth, errors.New("Could not find DisplayClaims object in XSTS response")
	}
	XSTSxui, ok := XSTSdc["xui"].([]interface{})
	if !ok {
		return auth, errors.New("Could not find xui array in DisplayClaims object")
	}
	if len(XSTSxui) < 1 {
		return auth, errors.New("xui array in DisplayClaims object does not have any elements")
	}
	XSTSuhsObject, ok := XSTSxui[0].(map[string]interface{})
	if !ok {
		return auth, errors.New("Could not get ush object in xui array")
	}
	XSTSuhs, ok := XSTSuhsObject["uhs"].(string)
	if !ok {
		return auth, errors.New("Could not get uhs string from ush object")
	}
	auth.UHS = XSTSuhs
	return auth, nil
}

// MCauth Represents Minecraft auth response
type MCauth struct {
	Token        string
	ExpiresAfter int64
}

// AuthMC Gets Minecraft authorization from XSTS token
func AuthMC(token XSTSauth) (MCauth, error) {
	var auth MCauth
	MCdataMap := map[string]interface{}{
		"identityToken": "XBL3.0 x=" + token.UHS + ";" + token.Token,
	}
	MCdata, err := json.Marshal(MCdataMap)
	if err != nil {
		return auth, err
	}
	MCreq, err := http.NewRequest(http.MethodPost, "https://api.minecraftservices.com/authentication/login_with_xbox", bytes.NewBuffer(MCdata))
	if err != nil {
		return auth, err
	}
	MCreq.Header.Set("Content-Type", "application/json")
	MCreq.Header.Set("Accept", "application/json")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	MCresp, err := client.Do(MCreq)
	if err != nil {
		return auth, err
	}
	var MCres map[string]interface{}
	json.NewDecoder(MCresp.Body).Decode(&MCres)
	MCresp.Body.Close()
	if MCresp.StatusCode != 200 {
		return auth, fmt.Errorf("MC answered not HTTP200! Instead got %s and following json: %#v", MCresp.Status, MCres)
	}
	MCtoken, ok := MCres["access_token"].(string)
	if !ok {
		return auth, errors.New("Could not find access_token in MC response")
	}
	auth.Token = MCtoken
	MCexpire, ok := MCres["expires_in"].(float64)
	if !ok {
		return auth, errors.New("Could not find expires_in in MC response")
	}
	auth.ExpiresAfter = time.Now().Unix() + int64(MCexpire)
	return auth, nil
}

// GetMCprofile Gets bot.Auth from token
func GetMCprofile(token string) (*AccountInfo, error) {
	var profile AccountInfo
	PRreq, err := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	if err != nil {
		return nil, err
	}
	PRreq.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	PRresp, err := client.Do(PRreq)
	if err != nil {
		return nil, err
	}
	var PRres map[string]interface{}
	json.NewDecoder(PRresp.Body).Decode(&PRres)
	PRresp.Body.Close()
	if PRresp.StatusCode != 200 {
		return nil, fmt.Errorf("MC (profile) answered not HTTP200! Instead got %s and following json: %#v", PRresp.Status, PRres)
	}
	PRuuid, ok := PRres["id"].(string)
	if !ok {
		return nil, errors.New("Could not find uuid in profile response")
	}
	profile.UUID = PRuuid
	PRname, ok := PRres["name"].(string)
	if !ok {
		return nil, errors.New("Could not find username in profile response")
	}
	profile.Username = PRname
	return &profile, nil
}

// GetMCcredentials From 0 to Minecraft bot.Auth with cache using device code flow
func GetMCcredentials(MSa MSauth) (*AccountInfo, error) {
	XBLa, err := AuthXBL(MSa.AccessToken)
	if err != nil {
		return nil, err
	}

	XSTSa, err := AuthXSTS(XBLa)
	if err != nil {
		return nil, err
	}

	MCa, err := AuthMC(XSTSa)
	if err != nil {
		return nil, err
	}

	auth, err := GetMCprofile(MCa.Token)
	if err != nil {
		return nil, err
	}
	auth.AccessToken = MCa.Token
	return auth, nil
}
