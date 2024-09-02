package notifiers

import (
	"errors"
	"log/slog"

	"github.com/hibare/GoS3Backup/internal/config"
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
		slog.Error("error running prechecks", "error", err)
		return
	}

	discordNotifyBackupSuccess(directory, totalDirs, totalFiles, successFiles, key)

}

func NotifyBackupFailure(directory string, totalDirs, totalFiles int, err error) {
	if err := runPreChecks(); err != nil {
		slog.Error("error running prechecks", "error", err)
		return
	}

	discordNotifyBackupFailure(directory, totalDirs, totalFiles, err)

}

func NotifyBackupDeleteFailure(key string, err error) {
	if err := runPreChecks(); err != nil {
		slog.Error("error running prechecks", "error", err)
		return
	}

	discordNotifyBackupDeleteFailure(key, err)
}
