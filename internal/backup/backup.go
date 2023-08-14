package backup

import (
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"

	commonDateTimes "github.com/hibare/GoCommon/pkg/datetime"
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/hibare/GoS3Backup/internal/notifiers"
	"github.com/hibare/GoS3Backup/internal/s3"
	"github.com/hibare/GoS3Backup/internal/utils"
)

func Backup() {
	sess := s3.NewSession(config.Current.S3)
	prefix := utils.GetTimeStampedPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname)
	log.Infof("prefix: %s", prefix)

	var uploadFunc func(*session.Session, string, string, string) (int, int, int)

	if config.Current.Backup.ArchiveDirs {
		log.Info("Archiving dirs")
		uploadFunc = s3.UploadZip
	} else {
		uploadFunc = s3.Upload
	}

	// Loop through individual backup dir & perform backup
	for _, dir := range config.Current.Backup.Dirs {
		log.Infof("Processing path %s", dir)

		totalFiles, totalDirs, successFiles := uploadFunc(sess, config.Current.S3.Bucket, prefix, dir)

		if successFiles <= 0 {
			log.Warnf("Uploaded files %d/%d", successFiles, totalFiles)
			notifiers.BackupFailedNotification("", dir, totalDirs, totalFiles)
			continue
		}

		notifiers.BackupSuccessfulNotification(dir, totalDirs, totalFiles, successFiles, prefix)
	}
	log.Info("Backup job ran successfully")
}

func ListBackups() []string {
	var keys []string
	sess := s3.NewSession(config.Current.S3)
	prefix := utils.GetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname)
	log.Infof("prefix: %s", prefix)

	// Retrieve objects by prefix
	keys, err := s3.ListObjectsAtPrefixRoot(sess, config.Current.S3.Bucket, prefix)
	if err != nil {
		log.Errorf("Error listing objects: %v", err)
		notifiers.BackupDeletionFailureNotification(err.Error(), constants.NotAvailable)
		return keys
	}

	if len(keys) == 0 {
		log.Info("No backups found")
		return keys
	}

	log.Infof("Found %d backups", len(keys))

	// Remove prefix from key to get datetime string
	keys = utils.TrimPrefix(keys, prefix)

	// Sort datetime strings by descending order
	sortedKeys := commonDateTimes.SortDateTimes(keys)

	return sortedKeys
}

func PurgeOldBackups() {
	sess := s3.NewSession(config.Current.S3)
	prefix := utils.GetPrefix(config.Current.S3.Prefix, config.Current.Backup.Hostname)

	backups := ListBackups()

	if len(backups) <= int(config.Current.Backup.RetentionCount) {
		log.Info("No backups to delete")
		return
	}

	keysToDelete := backups[config.Current.Backup.RetentionCount:]
	log.Infof("Found %d backups to delete (backup rentention %d) [%s]", len(keysToDelete), config.Current.Backup.RetentionCount, keysToDelete)

	// Delete datetime keys from S3 exceding retention count
	for _, key := range keysToDelete {
		log.Infof("Deleting backup %s", key)
		key = filepath.Join(prefix, key)

		if err := s3.DeleteObjects(sess, config.Current.S3.Bucket, key, true); err != nil {
			log.Errorf("Error deleting backup %s: %v", key, err)
			notifiers.BackupDeletionFailureNotification(err.Error(), key)
			continue
		}
	}

	log.Info("Deletion completed successfully")
}
