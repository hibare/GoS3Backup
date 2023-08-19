package backup

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	commonDateTimes "github.com/hibare/GoCommon/v2/pkg/datetime"
	commonFiles "github.com/hibare/GoCommon/v2/pkg/file"
	commonS3 "github.com/hibare/GoCommon/v2/pkg/s3"
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/constants"
	"github.com/hibare/GoS3Backup/internal/notifiers"
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
		log.Fatalf("Error creating session: %v", err)
		return
	}

	// Loop through individual backup dir & perform backup
	for _, dir := range config.Current.Backup.Dirs {
		log.Infof("Processing path %s", dir)

		if config.Current.Backup.ArchiveDirs {
			log.Infof("Archiving dir %s", dir)
			zipPath, totalFiles, totalDirs, successFiles, err := commonFiles.ArchiveDir(dir)
			if err != nil {
				log.Warnf("Archiving failed %s", dir)
				notifiers.BackupFailedNotification("", dir, totalDirs, totalFiles)
				continue
			}

			if successFiles <= 0 {
				err := fmt.Errorf("Failed to archive")
				log.Warnf("Uploading failed %s: %s", dir, err)
				notifiers.BackupFailedNotification(err.Error(), dir, totalDirs, totalFiles)
				continue
			}

			log.Infof("Uploading files %d/%d", successFiles, totalFiles)
			key, err := s3.UploadFile(zipPath)

			if err != nil {
				log.Warnf("Uploading failed %s: %s", dir, err)
				notifiers.BackupFailedNotification(err.Error(), dir, totalDirs, totalFiles)
				continue
			}

			log.Warnf("Uploaded files %d/%d at %s", successFiles, totalFiles, key)
			notifiers.BackupSuccessfulNotification(dir, totalDirs, totalFiles, successFiles, key)
			os.Remove(zipPath)

		} else {
			log.Infof("Uploading dir %s", dir)
			key, totalFiles, totalDirs, successFiles := s3.UploadDir(dir)

			if successFiles <= 0 {
				log.Warnf("Uploading failed %s", dir)
				notifiers.BackupFailedNotification("", dir, totalDirs, totalFiles)
				continue
			}

			log.Warnf("Uploaded files %d/%d at %s", successFiles, totalFiles, s3.Prefix)
			notifiers.BackupSuccessfulNotification(dir, totalDirs, totalFiles, successFiles, key)
		}

	}
	log.Info("Backup job ran successfully")
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
		log.Fatalf("Error creating session: %v", err)
		return keys, err
	}

	log.Infof("prefix: %s", s3.Prefix)

	// Retrieve objects by prefix
	keys, err := s3.ListObjectsAtPrefixRoot()
	if err != nil {
		log.Errorf("Error listing objects: %v", err)
		return keys, err
	}

	if len(keys) == 0 {
		log.Info("No backups found")
		return keys, nil
	}

	log.Infof("Found %d backups", len(keys))

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
		log.Fatalf("Error creating session: %v", err)
	}

	backups, err := ListBackups()
	if err != nil {
		notifiers.BackupDeletionFailureNotification(err.Error(), constants.NotAvailable)
		return
	}

	if len(backups) <= int(config.Current.Backup.RetentionCount) {
		log.Info("No backups to delete")
		return
	}

	keysToDelete := backups[config.Current.Backup.RetentionCount:]
	log.Infof("Found %d backups to delete (backup rentention %d) [%s]", len(keysToDelete), config.Current.Backup.RetentionCount, keysToDelete)

	// Delete datetime keys from S3 exceding retention count
	for _, key := range keysToDelete {
		log.Infof("Deleting backup %s", key)
		key = filepath.Join(s3.Prefix, key)

		if err := s3.DeleteObjects(key, true); err != nil {
			log.Errorf("Error deleting backup %s: %v", key, err)
			notifiers.BackupDeletionFailureNotification(err.Error(), key)
			continue
		}
	}

	log.Info("Deletion completed successfully")
}
