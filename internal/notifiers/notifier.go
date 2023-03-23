package notifiers

import (
	"github.com/hibare/GoS3Backup/internal/config"
	log "github.com/sirupsen/logrus"
)

func BackupSuccessfulNotification(directory string, dirs, files int, key string) {

	if !config.Current.Notifiers.Enabled {
		log.Warn("Notifiers are disabled")
		return
	}

	if config.Current.Notifiers.Discord.Webhook != "" && !config.Current.Notifiers.Discord.Enabled {
		log.Warning("Discord notifier not enabled")
		return
	} else if config.Current.Notifiers.Discord.Enabled {
		if err := DiscordBackupSuccessfulNotification(config.Current.Notifiers.Discord.Webhook, config.Current.Backup.Hostname, directory, dirs, files, key); err != nil {
			log.Errorf("Error sending Discord notification: %v", err)
		}
	}

}

func BackupFailedNotification(err, directory string, dirs, files int) {

	if !config.Current.Notifiers.Enabled {
		log.Warn("Notifiers are disabled")
		return
	}

	if config.Current.Notifiers.Discord.Webhook != "" && !config.Current.Notifiers.Discord.Enabled {
		log.Warning("Discord notifier not enabled")
		return
	} else if config.Current.Notifiers.Discord.Enabled {
		if err := DiscordBackupFailedNotification(config.Current.Notifiers.Discord.Webhook, config.Current.Backup.Hostname, err, directory, dirs, files); err != nil {
			log.Errorf("Error sending Discord notification: %v", err)
		}
	}

}

func BackupDeletionFailureNotification(err, key string) {

	if !config.Current.Notifiers.Enabled {
		log.Warn("Notifiers are disabled")
		return
	}

	if config.Current.Notifiers.Discord.Webhook != "" && !config.Current.Notifiers.Discord.Enabled {
		log.Warning("Discord notifier not enabled")
		return
	} else if config.Current.Notifiers.Discord.Enabled {
		if err := DiscordBackupDeletionFailureNotification(config.Current.Notifiers.Discord.Webhook, config.Current.Backup.Hostname, err, key); err != nil {
			log.Errorf("Error sending Discord notification: %v", err)
		}
	}
}
