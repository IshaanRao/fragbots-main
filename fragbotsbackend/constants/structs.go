package constants

// Bot Info Types

type BotInfo struct {
	BotId    string       `json:"botId" bson:"botId"`
	BotType  Bot          `json:"botType" bson:"botType"`
	ServerId string       `json:"serverId,omitempty" bson:"serverId,omitempty"`
	AccInfo  *AccountInfo `json:"accountInfo,omitempty" bson:"accInfo"`
	DiscInfo *DiscordInfo `json:"discordInfo" bson:"discInfo"`
	Running  bool         `json:"running,omitempty" bson:"running,omitempty"`
}

type DiscordInfo struct {
	LogWebhook     string `json:"logWebhook"`
	ConsoleWebhook string `json:"consoleWebhook"`
}

type AccountInfo struct {
	UUID        string `json:"uuid" bson:"uuid"`
	Username    string `json:"username" bson:"username"`
	Email       string `json:"email" bson:"email"`
	Password    string `json:"password" bson:"password"`
	AccessToken string `json:"accessToken" bson:"accessToken"`
	AuthData    MSauth `json:"msAuth" bson:"authData"`
}

// MSauth holds Microsoft auth credentials
type MSauth struct {
	AccessToken  string `json:"accessToken" bson:"accessToken"`
	ExpiresAfter int64  `json:"expiresAfter" bson:"expiresAfter"`
	RefreshToken string `json:"refreshToken" bson:"refreshToken"`
}

type Bot string

const (
	Priority    Bot = "PRIORITY"
	Exclusive       = "EXCLUSIVE"
	Active          = "ACTIVE"
	Whitelisted     = "WHITELISTED"
	Verified        = "VERIFIED"
)
