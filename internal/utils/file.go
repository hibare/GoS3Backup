package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func ListFilesDirs(root string, exclude []*regexp.Regexp) ([]string, []string) {
	var files []string
	var dirs []string

	readDir := func(dir string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Check if directory matches any of the exclude patterns
			for _, e := range exclude {
				if e.MatchString(d.Name()) {
					return filepath.SkipDir
				}
			}

			dirs = append(dirs, filepath.Join(dir, d.Name()))
		} else {
			// Check if file matches any of the exclude patterns
			for _, e := range exclude {
				if e.MatchString(d.Name()) {
					return nil
				}
			}

			files = append(files, filepath.Join(dir, d.Name()))
		}

		return nil
	}

	_ = filepath.WalkDir(root, readDir)

	return files, dirs
}

func ZipDir(dirPath string) (error, string, int, int, int) {
	dirPath = filepath.Clean(dirPath)
	dirName := filepath.Base(dirPath)
	zipName := fmt.Sprintf("%s.zip", dirName)
	zipPath := filepath.Join(os.TempDir(), zipName)
	totalFiles, totalDirs, successFiles := 0, 0, 0

	// Create a temporary file to hold the zip archive
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err, zipPath, totalFiles, totalDirs, successFiles
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			totalDirs++
			return nil
		}

		totalFiles++

		file, err := os.Open(path)
		if err != nil {
			log.Errorf("Failed to open file: %v", err)
			return nil
		}
		defer file.Close()

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			log.Errorf("Failed to get relative path: %v", err)
			return nil
		}

		info, err := d.Info()
		if err != nil {
			log.Errorf("Failed to get file info: %v", err)
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Errorf("Failed to create header: %v", err)
			return nil
		}
		header.Name = filepath.ToSlash(filepath.Join(relPath))

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			log.Errorf("Failed to create header: %v", err)
			return nil
		}

		_, err = io.Copy(writer, file)
		if err != nil {
			log.Errorf("Failed to write file to archive: %v", err)
			return nil
		}
		successFiles++

		return nil
	})

	log.Infof("Created archive '%s' for directory '%s'", zipPath, dirPath)
	return err, zipPath, totalFiles, totalDirs, successFiles

}
