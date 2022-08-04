package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"os"
)

// General bot info

const Name = "FragBot"
const GuildId = "816493066910302240"

var DebugMode = true
var Token string
var AccessToken string
var BackendUrl string

// Permissions

var StaffPerms int64 = discordgo.PermissionManageMessages
var ModPerms int64 = discordgo.PermissionManageRoles
var AdminPerms int64 = discordgo.PermissionAdministrator

// Embeds
const FooterIcon = "https://cdn.discordapp.com/emojis/823999418592264232.webp?size=240&quality=lossless"
const FooterText = "FragBots"
const DefaultEmbedColor = 15954943

func init() {
	if err := godotenv.Load(); err != nil {
		LogFatal("No ..env file found")
	}

	Token = getEnv("TOKEN")
	AccessToken = getEnv("ACCESS_TOKEN")
	BackendUrl = getEnv("BACKEND_URI")
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		LogFatal("No" + key + "found in env")
	}
	return val
}
