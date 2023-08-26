package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"jsfraz/mega-backuper/models"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Create tarball and return path to it.
//
//	@param backup
//	@return string Path to tarball.
//	@return string Tarball file name.
//	@return error
func createTarball(backup models.Backup, now time.Time) (string, string, error) {
	// create tarball file
	folderPath := "/tmp/" + backup.Name
	tarballFileName := backup.Name + "_" + now.Format(time.RFC3339) + ".tar.gz"
	tarballPath := "/tmp/" + tarballFileName
	tarballFile, err := os.Create(tarballPath)
	if err != nil {
		return "", "", err
	}
	defer tarballFile.Close()
	// create a GZIP writer to compress the tarball
	gzipWriter := gzip.NewWriter(tarballFile)
	defer gzipWriter.Close()

	// create a TAR writer to write files and headers to the compressed stream
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// walk through the source folder and add its contents to the tarball
	err = filepath.Walk(folderPath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// create a TAR header based on the file info
		header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return err
		}

		// calculate the relative path of the file within the source folder
		relPath, _ := filepath.Rel(folderPath, filePath)
		header.Name = relPath

		// write the header to the TAR archive
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// if the file is a regular file, copy its content to the TAR archive
		if fileInfo.Mode().IsRegular() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			// copy the file's content to the TAR archive
			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})

	// return result
	if err != nil {
		return "", "", err
	} else {
		return tarballPath, tarballFileName, nil
	}
}

// Get /each/path/element/ from path string.
//
//	@param path
//	@return []string
func extractPaths(path string) []string {
	substrings := strings.Split(path, "/")
	// remove empty strings from the resulting slice
	var result []string
	for _, s := range substrings {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// Uploads file to Mega and deletes it locally.
//
//	@param localFilePath
//	@param fileName
//	@param megaDir
//	@return error
func uploadToMegaAndDelete(localFilePath string, fileName string, megaDir string) error {
	// check mega dir and return node to upload to
	uploadNode, err := MegaCheckDir(megaDir)
	if err != nil {
		return err
	}
	// upload
	err = MegaUpload(localFilePath, uploadNode, fileName)
	if err != nil {
		return err
	}
	// delete file
	removeErr := os.Remove(localFilePath)
	if removeErr != nil {
		log.Println("Error deleting "+localFilePath+": ", removeErr)
	}
	return err
}

// Backup volume to Mega.
//
//	@param backup
//	@return error
func BackupVolume(backup models.Backup) error {
	now := time.Now()
	// make tarball
	tarballPath, tarballFileName, err := createTarball(backup, now)
	if err != nil {
		return err
	}
	// upload to mega
	err = uploadToMegaAndDelete(tarballPath, tarballFileName, backup.MegaDir)
	return err
	// TODO delete oldest file(s)
}
