package main

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
