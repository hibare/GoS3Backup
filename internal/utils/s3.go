package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/hibare/GoS3Backup/internal/constants"
)

func GetPrefix(prefix string, hostname string) string {
	prefixSlice := []string{}
	if prefix != "" {
		prefixSlice = append(prefixSlice, prefix)
	}

	if hostname != "" {
		prefixSlice = append(prefixSlice, hostname)
	}

	generatedPrefix := filepath.Join(prefixSlice...)

	if !strings.HasSuffix(generatedPrefix, constants.PrefixSeparator) {
		generatedPrefix = fmt.Sprintf("%s%s", generatedPrefix, constants.PrefixSeparator)
	}

	return generatedPrefix
}
func GetTimeStampedPrefix(prefix string, hostname string) string {
	timePrefix := time.Now().Format(constants.DefaultDateTimeLayout)
	prefix = GetPrefix(prefix, hostname)

	generatedPrefix := filepath.Join(prefix, timePrefix)

	if !strings.HasSuffix(generatedPrefix, constants.PrefixSeparator) {
		generatedPrefix = fmt.Sprintf("%s%s", generatedPrefix, constants.PrefixSeparator)
	}

	return generatedPrefix

}

func TrimPrefix(keys []string, prefix string) []string {
	var trimmedKeys []string
	for _, key := range keys {
		trimmedKey := strings.TrimPrefix(key, prefix)
		trimmedKey = strings.TrimSuffix(trimmedKey, "/")
		trimmedKeys = append(trimmedKeys, trimmedKey)
	}
	return trimmedKeys
}
