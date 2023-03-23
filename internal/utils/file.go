package utils

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func EstimateFilesAndDirs(path string) (int, int) {
	var files, dirs int

	filesInfo, err := os.ReadDir(path)
	if err != nil {
		log.Errorf("Error reading directory: %v", err)
		return 0, 0
	}

	// Recursively traverse directory to count files & dirs.
	for _, file := range filesInfo {
		if file.IsDir() {
			dirs++
			subdirPath := filepath.Join(path, file.Name())
			log.Infof("Found directory: %s", subdirPath)
			subFiles, subDirs := EstimateFilesAndDirs(subdirPath)
			if err != nil {
				return subFiles, subDirs
			}
			files += subFiles
			dirs += subDirs
		} else {
			files++
		}
	}

	return files, dirs
}
