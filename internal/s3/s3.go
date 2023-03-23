package s3

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hibare/GoS3Backup/internal/config"
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

func Upload(sess *session.Session, bucket, prefix, baseDir string) error {
	client := s3.New(sess)
	baseDirParentPath := filepath.Dir(baseDir)

	// Recursively upload files preserving paths w.r.t prefix
	// We are not returning error, and continue to with next directory incase of failure
	filepath.WalkDir(baseDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Errorf("Error walking path %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Errorf("Error opening file %s: %v", path, err)
			return nil
		}
		defer file.Close()

		key := filepath.Join(prefix, strings.TrimPrefix(path, baseDirParentPath))
		_, err = client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   file,
		})
		if err != nil {
			log.Errorf("Error uploading file %s: %v", path, err)
			return nil
		}
		log.Infof("Uploaded %s to S3://%s/%s", path, bucket, key)
		return nil
	})

	return nil
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
