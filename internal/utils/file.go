package utils

import (
	"archive/zip"
	"fmt"
	"io"
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

func ZipDir(dirPath string) (error, string, int, int, int) {
	dirName := filepath.Base(dirPath)
	zipName := fmt.Sprintf("%s.zip", dirName)
	zipPath := filepath.Join(os.TempDir(), zipName)
	totalFiles, totalDirs, successFiles := 0, 0, 0

	files, dirs := ListFilesDirs(dirPath)
	totalFiles = len(files)
	totalDirs = len(dirs)
	log.Infof("Found %d files from %d directories", totalFiles, totalDirs)

	// Create a temporary file to hold the zip archive
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err, zipPath, totalFiles, totalDirs, successFiles
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add each file to the zip archive
	for _, file := range files {
		// Open the file
		fileToZip, err := os.Open(file)
		if err != nil {
			log.Errorf("Failed to open file: %v", err)
			continue
		}
		defer fileToZip.Close()

		// Get the file info to create the zip header
		info, err := fileToZip.Stat()
		if err != nil {
			log.Errorf("Failed to get file info: %v", err)
			continue
		}

		// Create a new zip file header for the current file
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Errorf("Failed to create header: %v", err)
			continue
		}

		// Set the name of the file relative to the base directory to include in the archive
		relPath, err := filepath.Rel(dirPath, file)
		if err != nil {
			log.Errorf("Failed to get relative path: %v", err)
			continue
		}
		header.Name = relPath

		// Add the header to the zip archive
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			log.Errorf("Failed to create header: %v", err)
			continue
		}

		// Copy the contents of the file to the zip archive
		_, err = io.Copy(writer, fileToZip)
		if err != nil {
			log.Errorf("Failed to write file to archive: %v", err)
			continue
		}
		successFiles++
	}

	log.Infof("Created archive '%s' for directory '%s'", zipPath, dirPath)
	return nil, zipPath, totalFiles, totalDirs, successFiles

}
