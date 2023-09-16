package notifiers

import (
	"fmt"
	"strconv"

	"github.com/hibare/GoCommon/v2/pkg/notifiers/discord"
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/hibare/GoS3Backup/internal/version"
	"github.com/rs/zerolog/log"
)

func runDiscordPrechecks() error {
	if !config.Current.Notifiers.Discord.Enabled {
		return ErrNotifierDisabled
	}
	return nil
}

func discordNotifyBackupSuccess(directory string, totalDirs, totalFiles, successFiles int, key string) {
	if err := runDiscordPrechecks(); err != nil {
		log.Error().Err(err).Msg("error running discord prechecks")
		return
	}

	message := discord.Message{
		Embeds: []discord.Embed{
			{
				Title:       "Directory",
				Description: directory,
				Color:       1498748,
				Fields: []discord.EmbedField{
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
		Components: []discord.Component{},
		Username:   constants.ProgramIdentifier,
		Content:    fmt.Sprintf("**Backup Successful** - *%s*", config.Current.Backup.Hostname),
	}

	if version.V.NewVersionAvailable {
		if err := message.AddFooter(version.V.GetUpdateNotification()); err != nil {
			log.Error().Err(err)
		}
	}

	if err := message.Send(config.Current.Notifiers.Discord.Webhook); err != nil {
		log.Error().Err(err)
	}
}

func discordNotifyBackupFailure(directory string, totalDirs, totalFiles int, err error) {
	if err := runDiscordPrechecks(); err != nil {
		log.Error().Err(err)
		return
	}

	message := discord.Message{
		Embeds: []discord.Embed{
			{
				Title:       "Error",
				Description: err.Error(),
				Color:       14554702,
				Fields: []discord.EmbedField{
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
		Components: []discord.Component{},
		Username:   constants.ProgramIdentifier,
		Content:    fmt.Sprintf("**Backup Failed** - *%s*", config.Current.Backup.Hostname),
	}

	if version.V.NewVersionAvailable {
		if err := message.AddFooter(version.V.GetUpdateNotification()); err != nil {
			log.Error().Err(err)
		}
	}

	if err := message.Send(config.Current.Notifiers.Discord.Webhook); err != nil {
		log.Error().Err(err)
	}
}

func discordNotifyBackupDeleteFailure(key string, err error) {
	if err := runDiscordPrechecks(); err != nil {
		log.Error().Err(err)
		return
	}

	message := discord.Message{
		Embeds: []discord.Embed{
			{
				Title:       "Error",
				Description: err.Error(),
				Color:       14590998,
				Fields: []discord.EmbedField{
					{
						Name:   "Key",
						Value:  key,
						Inline: false,
					},
				},
			},
		},
		Components: []discord.Component{},
		Username:   constants.ProgramIdentifier,
		Content:    fmt.Sprintf("**Backup Deletion Failed** - *%s*", config.Current.Backup.Hostname),
	}

	if version.V.NewVersionAvailable {
		if err := message.AddFooter(version.V.GetUpdateNotification()); err != nil {
			log.Error().Err(err)
		}
	}

	if err := message.Send(config.Current.Notifiers.Discord.Webhook); err != nil {
		log.Error().Err(err)
	}
}
