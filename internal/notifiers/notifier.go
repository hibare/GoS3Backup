package notifiers

import (
	"errors"

	"github.com/hibare/GoS3Backup/internal/config"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNotifiersDisabled      = errors.New("notifiers are disabled")
	ErrMissingNotifierWebhook = errors.New("missing notifier webhook")
	ErrNotifierDisabled       = errors.New("notifier is disabled")
)

func runPreChecks() error {
	if !config.Current.Notifiers.Enabled {
		return ErrNotifiersDisabled
	}

	return nil
}

func NotifyBackupSuccess(directory string, totalDirs, totalFiles, successFiles int, key string) {
	if err := runPreChecks(); err != nil {
		log.Error(err)
		return
	}

	discordNotifyBackupSuccess(directory, totalDirs, totalFiles, successFiles, key)

}

func NotifyBackupFailure(directory string, totalDirs, totalFiles int, err error) {
	if err := runPreChecks(); err != nil {
		log.Error(err)
		return
	}

	discordNotifyBackupFailure(directory, totalDirs, totalFiles, err)

}

func NotifyBackupDeleteFailure(key string, err error) {
	if err := runPreChecks(); err != nil {
		log.Error(err)
		return
	}

	discordNotifyBackupDeleteFailure(key, err)
}
