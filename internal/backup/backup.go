package backup

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

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
		log.Fatal().Err(err).Msg("Error creating session")
		return
	}

	// Loop through individual backup dir & perform backup
	for _, dir := range config.Current.Backup.Dirs {
		log.Info().Msgf("Processing path %s", dir)

		if config.Current.Backup.ArchiveDirs {
			log.Info().Msgf("Archiving dir %s", dir)
			archivePath, totalFiles, totalDirs, successFiles, err := commonFiles.ArchiveDir(dir, nil)
			if err != nil {
				log.Error().Err(err).Msgf("Archiving failed %s", dir)
				notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, err)
				continue
			}

			if successFiles <= 0 {
				log.Error().Err(err).Msgf("Uploading failed %s", dir)
				notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, ErrNoProcessableFiles)
				continue
			}
			log.Info().Msgf("Archived files %d/%d, archive path %s", successFiles, totalFiles, archivePath)

			uploadPath := archivePath

			if config.Current.Backup.Encryption.Enabled {
				log.Info().Msgf("Encrypting archive %s", archivePath)
				gpg, err := commonGPG.DownloadGPGPubKey(config.Current.Backup.Encryption.GPG.KeyID, config.Current.Backup.Encryption.GPG.KeyServer)
				if err != nil {
					log.Error().Err(err).Msg("Error downloading gpg key")
					notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, err)
					continue
				}

				encryptedFilePath, err := gpg.EncryptFile(archivePath)
				if err != nil {
					log.Error().Err(err).Msg("Error encrypting file")
					notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, err)
					continue
				}

				uploadPath = encryptedFilePath
				log.Info().Msgf("Archive encrypted at %s", encryptedFilePath)
				os.Remove(archivePath)
			}

			log.Info().Msgf("Uploading file %s", uploadPath)
			key, err := s3.UploadFile(uploadPath)
			if err != nil {
				log.Error().Err(err).Msgf("Uploading failed %s", dir)
				notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, err)
				continue
			}

			log.Info().Msgf("Uploaded files %d/%d at %s", successFiles, totalFiles, key)
			notifiers.NotifyBackupSuccess(dir, totalDirs, totalFiles, successFiles, key)
			os.Remove(uploadPath)
		} else {
			log.Info().Msgf("Uploading dir %s", dir)
			key, totalFiles, totalDirs, successFiles := s3.UploadDir(dir, nil)

			if successFiles <= 0 {
				log.Warn().Msgf("Uploading failed %s", dir)
				notifiers.NotifyBackupFailure(dir, totalDirs, totalFiles, ErrNoProcessableFiles)
				continue
			}

			log.Warn().Msgf("Uploaded files %d/%d at %s", successFiles, totalFiles, s3.Prefix)
			notifiers.NotifyBackupSuccess(dir, totalDirs, totalFiles, successFiles, key)
		}

	}
	log.Info().Msg("Backup job ran successfully")
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
		log.Error().Err(err).Msg("Error creating session")
		return keys, err
	}

	log.Info().Msgf("prefix: %s", s3.Prefix)

	// Retrieve objects by prefix
	keys, err := s3.ListObjectsAtPrefixRoot()
	if err != nil {
		log.Error().Err(err).Msg("Error listing objects")
		return keys, err
	}

	if len(keys) == 0 {
		log.Info().Msg("No backups found")
		return keys, nil
	}

	log.Info().Msgf("Found %d backups", len(keys))

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
		log.Error().Err(err).Msg("Error creating session")
	}

	backups, err := ListBackups()
	if err != nil {
		notifiers.NotifyBackupDeleteFailure(constants.NotAvailable, err)
		return
	}

	if len(backups) <= int(config.Current.Backup.RetentionCount) {
		log.Info().Msg("No backups to delete")
		return
	}

	keysToDelete := backups[config.Current.Backup.RetentionCount:]
	log.Info().Msgf("Found %d backups to delete (backup rentention %d) [%s]", len(keysToDelete), config.Current.Backup.RetentionCount, keysToDelete)

	// Delete datetime keys from S3 exceding retention count
	for _, key := range keysToDelete {
		log.Info().Msgf("Deleting backup %s", key)
		key = filepath.Join(s3.Prefix, key)

		if err := s3.DeleteObjects(key, true); err != nil {
			log.Error().Err(err).Msgf("Error deleting backup %s", key)
			notifiers.NotifyBackupDeleteFailure(key, err)
			continue
		}
	}

	log.Info().Msg("Deletion completed successfully")
}
