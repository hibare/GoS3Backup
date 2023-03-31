package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/version"
	log "github.com/sirupsen/logrus"
)

type DiscordWebhookMessage struct {
	Embeds     []DiscordEmbed     `json:"embeds"`
	Components []DiscordComponent `json:"components"`
	Username   string             `json:"username"`
	Content    string             `json:"content"`
}

type DiscordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Footer      DiscordEmbedFooter  `json:"footer"`
	Fields      []DiscordEmbedField `json:"fields"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordEmbedFooter struct {
	Text string `json:"text"`
}

type DiscordComponent struct {
	// Define struct for Discord components if needed
}

func (d *DiscordWebhookMessage) AddFooter() {
	isUpdate, LatestVersion := version.IsNewVersionAvailable()
	if isUpdate {
		footer := DiscordEmbedFooter{
			Text: fmt.Sprintf(version.UpdateNotificationMessage, LatestVersion),
		}
		d.Embeds[0].Footer = footer
	}
}

func DiscordBackupSuccessfulNotification(webhookUrl string, hostname, directory string, totalDirs, totalFiles, successFiles int, key string) error {
	webhookMessage := DiscordWebhookMessage{
		Embeds: []DiscordEmbed{
			{
				Title:       "Directory",
				Description: directory,
				Color:       1498748,
				Fields: []DiscordEmbedField{
					{
						Name:   "Key",
						Value:  key,
						Inline: false,
					},
					{
						Name:   "Dirs",
						Value:  strconv.Itoa(totalDirs),
						Inline: true,
					},
					{
						Name:   "Files",
						Value:  fmt.Sprintf("%d/%d", successFiles, totalFiles),
						Inline: true,
					},
				},
			},
		},
		Components: []DiscordComponent{},
		Username:   "Backup Job",
		Content:    fmt.Sprintf("**Backup Successful** - *%s*", hostname),
	}
	webhookMessage.AddFooter()

	return SendMessage(webhookUrl, webhookMessage)

}

func DiscordBackupFailedNotification(webhookUrl string, hostname, err, directory string, totalDirs, totalFiles int) error {
	webhookMessage := DiscordWebhookMessage{
		Embeds: []DiscordEmbed{
			{
				Title:       "Error",
				Description: err,
				Color:       14554702,
				Fields: []DiscordEmbedField{
					{
						Name:   "Directory",
						Value:  directory,
						Inline: false,
					},
					{
						Name:   "Dirs",
						Value:  strconv.Itoa(totalDirs),
						Inline: true,
					},
					{
						Name:   "Files",
						Value:  strconv.Itoa(totalFiles),
						Inline: true,
					},
				},
			},
		},
		Components: []DiscordComponent{},
		Username:   "Backup Job",
		Content:    fmt.Sprintf("**Backup Failed** - *%s*", hostname),
	}
	webhookMessage.AddFooter()

	return SendMessage(webhookUrl, webhookMessage)
}

func DiscordBackupDeletionFailureNotification(webhookUrl string, hostname, err, key string) error {
	webhookMessage := DiscordWebhookMessage{
		Embeds: []DiscordEmbed{
			{
				Title:       "Error",
				Description: err,
				Color:       14590998,
				Fields: []DiscordEmbedField{
					{
						Name:   "Key",
						Value:  key,
						Inline: false,
					},
				},
			},
		},
		Components: []DiscordComponent{},
		Username:   "Backup Job",
		Content:    fmt.Sprintf("**Backup Deletion Failed** - *%s*", hostname),
	}
	webhookMessage.AddFooter()

	return SendMessage(webhookUrl, webhookMessage)
}

func SendMessage(webhookUrl string, message DiscordWebhookMessage) error {
	if config.Current.Notifiers.Discord.Webhook != "" && !config.Current.Notifiers.Discord.Enabled {
		log.Warning("Discord notifier not enabled")
		return nil
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return &json.SyntaxError{}
	}

	resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
