package backup

import (
	"errors"
	"os"
	"path/filepath"

	"log/slog"

	commonGPG "github.com/hibare/GoCommon/v2/pkg/crypto/gpg"
	commonDateTimes "github.com/hibare/GoCommon/v2/pkg/datetime"
	commonFiles "github.com/hibare/GoCommon/v2/pkg/file"
	commonS3 "github.com/hibare/GoCommon/v2/pkg/s3"
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/hibare/GoS3Backup/internal/notifiers"
)

var (
	ErrArchiving          = errors.New("error archiving")
	ErrNoProcessableFiles = errors.New("no processable files")
)

func Backup() {
	s3 := commonS3.S3{
		Endpoint:  config.Current.S3.Endpoint,
		Region:    config.Current.S3.Region,
		AccessKey: config.Current.S3.AccessKey,
		SecretKey: config.Current.S3.SecretKey,
		Bucket:    config.Current.S3.Bucket,
	}

	s3.SetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname, true)

	if err := s3.NewSession(); err != nil {
		slog.Error("Error creating session", "error", err)
		os.Exit(1)
		return
	}

	// Loop through individual backup dir & perform backup
	for _, dir := range config.Current.Backup.Dirs {
		slog.Info("Processing path", "path", dir)

		if config.Current.Backup.ArchiveDirs {
			slog.Info("Archiving dir", "dir", dir)
			archivePath, totalFiles, totalDirs, successFiles, err := commonFiles.ArchiveDir(dir, nil)
			if err != nil {
				slog.Error("Error archiving", "error", err)
				notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, err)
				continue
			}

			if successFiles <= 0 {
				slog.Error("No processable files", "dir", dir)
				notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, ErrNoProcessableFiles)
				continue
			}
			slog.Info("Archived files", "successFiles", successFiles, "totalFiles", totalFiles, "archivePath", archivePath)

			uploadPath := archivePath

			if config.Current.Backup.Encryption.Enabled {
				slog.Info("Encrypting archive", "archivePath", archivePath)
				gpg, err := commonGPG.DownloadGPGPubKey(config.Current.Backup.Encryption.GPG.KeyID, config.Current.Backup.Encryption.GPG.KeyServer)
				if err != nil {
					slog.Error("Error downloading gpg key", "error", err)
					notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, err)
					continue
				}

				encryptedFilePath, err := gpg.EncryptFile(archivePath)
				if err != nil {
					slog.Error("Error encrypting file", "error", err)
					notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, err)
					continue
				}

				uploadPath = encryptedFilePath
				slog.Info("Encrypted archive", "uploadPath", uploadPath)
				os.Remove(archivePath)
			}

			slog.Info("Uploading file", "uploadPath", uploadPath)
			key, err := s3.UploadFile(uploadPath)
			if err != nil {
				slog.Error("Uploading failed", "error", err)
				notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, err)
				continue
			}

			slog.Info("Uploaded file", "key", key, "successFiles", successFiles, "totalFiles", totalFiles, "uploadPath", uploadPath)
			notifiers.NotifyBackupSuccess(dir, totalDirs, totalFiles, successFiles, key)
			os.Remove(uploadPath)
		} else {
			slog.Info("Uploading dir", "dir", dir)
			key, totalFiles, totalDirs, successFiles := s3.UploadDir(dir, nil)

			if successFiles <= 0 {
				slog.Warn("No processable files", "dir", dir)
				notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, ErrNoProcessableFiles)
				continue
			}

			slog.Warn("Uploaded files", "successFiles", successFiles, "totalFiles", totalFiles, "dir", dir)
			notifiers.NotifyBackupSuccess(dir, totalDirs, totalFiles, successFiles, key)
		}

	}
	slog.Info("Backup job ran successfully")
}

func ListBackups() ([]string, error) {
	var keys []string

	s3 := commonS3.S3{
		Endpoint:  config.Current.S3.Endpoint,
		Region:    config.Current.S3.Region,
		AccessKey: config.Current.S3.AccessKey,
		SecretKey: config.Current.S3.SecretKey,
		Bucket:    config.Current.S3.Bucket,
	}

	s3.SetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname, false)

	if err := s3.NewSession(); err != nil {
		slog.Error("Error creating session", "error", err)
		return keys, err
	}

	slog.Info("prefix", "prefix", s3.Prefix)

	// Retrieve objects by prefix
	keys, err := s3.ListObjectsAtPrefixRoot()
	if err != nil {
		slog.Error("Error listing objects", "error", err)
		return keys, err
	}

	if len(keys) == 0 {
		slog.Info("No backups found")
		return keys, nil
	}

	slog.Info("Found backups", "keys", len(keys))

	// Remove prefix from key to get datetime string
	keys = s3.TrimPrefix(keys)

	// Sort datetime strings by descending order
	sortedKeys := commonDateTimes.SortDateTimes(keys)

	return sortedKeys, nil
}

func PurgeOldBackups() {
	s3 := commonS3.S3{
		Endpoint:  config.Current.S3.Endpoint,
		Region:    config.Current.S3.Region,
		AccessKey: config.Current.S3.AccessKey,
		SecretKey: config.Current.S3.SecretKey,
		Bucket:    config.Current.S3.Bucket,
	}
	s3.SetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname, false)

	if err := s3.NewSession(); err != nil {
		slog.Error("Error creating session", "error", err)
	}

	backups, err := ListBackups()
	if err != nil {
		notifiers.NotifyBackupDeleteFailure(constants.NotAvailable, err)
		return
	}

	if len(backups) <= int(config.Current.Backup.RetentionCount) {
		slog.Info("No backups to delete")
		return
	}

	keysToDelete := backups[config.Current.Backup.RetentionCount:]
	slog.Info("Found backups to delete", "backups", len(keysToDelete), "retention", config.Current.Backup.RetentionCount, "keys", keysToDelete)

	// Delete datetime keys from S3 exceding retention count
	for _, key := range keysToDelete {
		slog.Info("Deleting backup", "key", key)
		key = filepath.Join(s3.Prefix, key)

		if err := s3.DeleteObjects(key, true); err != nil {
			slog.Error("Error deleting backup", "key", key, "error", err)
			notifiers.NotifyBackupDeleteFailure(key, err)
			continue
		}
	}

	slog.Info("Deletion completed successfully")
}
