package utils

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func ListFilesDirs(path string) ([]string, []string) {
	var (
		files []string
		Dirs  []string
	)

	filesInfo, err := os.ReadDir(path)
	if err != nil {
		log.Errorf("Error reading directory: %v", err)
		return files, Dirs
	}

	// Recursively traverse directory to list files & dirs.
	for _, file := range filesInfo {
		if file.IsDir() {
			subdirPath := filepath.Join(path, file.Name())
			subFiles, subDirs := ListFilesDirs(subdirPath)
			files = append(files, subFiles...)
			Dirs = append(Dirs, subdirPath)
			for _, dir := range subDirs {
				Dirs = append(Dirs, filepath.Join(path, dir))
			}
		} else {
			files = append(files, filepath.Join(path, file.Name()))
		}
	}

	return files, Dirs
}
