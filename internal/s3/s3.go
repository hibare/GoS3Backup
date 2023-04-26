package s3

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hibare/GoS3Backup/internal/config"
	"github.com/hibare/GoS3Backup/internal/utils"
	log "github.com/sirupsen/logrus"
)

func NewSession(s3Config config.S3Config) *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region:      &s3Config.Region,
		Endpoint:    &s3Config.Endpoint,
		Credentials: credentials.NewStaticCredentials(s3Config.AccessKey, s3Config.SecretKey, ""),
	})

	if err != nil {
		log.Fatalf("Error creating session: %v", err)
	}

	return sess
}

func Upload(sess *session.Session, bucket, prefix, baseDir string) (int, int, int) {
	totalFiles, totalDirs, successFiles := 0, 0, 0

	client := s3.New(sess)
	baseDirParentPath := filepath.Dir(baseDir)

	files, dirs := utils.ListFilesDirs(baseDir, nil)

	totalFiles = len(files)
	totalDirs = len(dirs)

	for _, file := range files {
		fp, err := os.Open(file)
		if err != nil {
			log.Errorf("Error opening file %s: %v", file, err)
			continue
		}
		defer fp.Close()

		key := filepath.Join(prefix, strings.TrimPrefix(file, baseDirParentPath))
		_, err = client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   fp,
		})
		if err != nil {
			log.Errorf("Error uploading file %s: %v", file, err)
			continue
		}
		successFiles += 1
		log.Infof("Uploaded %s to S3://%s/%s", file, bucket, key)
	}

	return totalFiles, totalDirs, successFiles
}

func UploadZip(sess *session.Session, bucket, prefix, baseDir string) (int, int, int) {
	totalFiles, totalDirs, successFiles := 0, 0, 0

	err, zipPath, totalFiles, totalDirs, successFiles := utils.ArchiveDir(baseDir)

	if err != nil {
		log.Errorf("Error creating zip file: %v", err)
		return totalFiles, totalDirs, 0
	}

	// Create an S3 uploader using the session
	uploader := s3manager.NewUploader(sess)

	// Open the file to upload
	f, err := os.Open(zipPath)
	if err != nil {
		log.Errorf("Failed to open file %v", err)
		return totalFiles, totalDirs, 0
	}
	defer f.Close()

	log.Infof("Uploading file %s to S3://%s/%s", zipPath, bucket, prefix)

	// Upload the file to S3
	key := filepath.Join(prefix, filepath.Base(zipPath))
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	})
	if err != nil {
		log.Errorf("Failed to upload file %v", err)
		return totalFiles, totalDirs, 0
	}
	log.Infof("Uploaded %s to S3://%s/%s", zipPath, bucket, key)

	if err = os.Remove(zipPath); err != nil {
		log.Errorf("Failed to remove zip file %v", err)
	}
	log.Infof("Removed zip file %s", zipPath)

	return totalFiles, totalDirs, successFiles
}

func ListObjectsAtPrefixRoot(sess *session.Session, bucket, prefix string) ([]string, error) {
	client := s3.New(sess)

	var keys []string
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}

	resp, err := client.ListObjectsV2(input)
	if err != nil {
		return keys, err
	}

	for _, obj := range resp.Contents {
		if *obj.Key == prefix {
			continue
		}
		keys = append(keys, *obj.Key)
	}

	if len(keys) == 0 && len(resp.CommonPrefixes) == 0 {
		return keys, nil
	}

	for _, cp := range resp.CommonPrefixes {
		keys = append(keys, *cp.Prefix)
	}

	return keys, nil
}

func DeleteObjects(sess *session.Session, bucket, key string, recursive bool) error {
	client := s3.New(sess)

	// Delete all child object recursively
	if recursive {
		log.Warnf("Recursively deleting objects in bucket S3://%s/%s", bucket, key)
		// List all objects in the bucket with the given key
		resp, err := client.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(bucket),
			Prefix: aws.String(key),
		})
		if err != nil {
			return err
		}

		log.Infof("Found %d objects in bucket S3://%s/%s", len(resp.Contents), bucket, key)

		// Delete all objects with the given key
		for _, obj := range resp.Contents {
			_, err = client.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(bucket),
				Key:    obj.Key,
			})

			if err != nil {
				return err
			}
			log.Infof("Deleted object with key '%s' from bucket '%s'", *obj.Key, bucket)
		}
	}

	// Delete the prefix
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return err
	}

	log.Infof("Deleted key %s", key)

	return nil
}
