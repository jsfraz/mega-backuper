package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"jsfraz/mega-backuper/models"
	"log"
	"os"
	"path/filepath"
)

// Create tarball and return path to it.
//
//	@param backup
//	@return string Path to tarball.
//	@return error
func CreateTarball(backup models.Backup) (string, error) {
	// create tarball file
	folderPath := "/tmp/" + backup.Name
	tarballPath := "/tmp/" + backup.Name + ".tar.gz"
	tarballFile, err := os.Create(tarballPath)
	if err != nil {
		return "", err
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
		return "", err
	} else {
		return tarballPath, nil
	}
}

// Backup volume.
//
//	@param backup
func BackupVolume(backup models.Backup) {
	log.Println("Backing up [" + string(backup.Type) + "] backup job '" + backup.Name + "'...")
	// make tarball
	tarballPath, err := CreateTarball(backup)
	if err != nil {
		log.Println("Failed to backup ["+string(backup.Type)+"] backup job '"+backup.Name+"': ", err)
	}
	// TODO upload to Mega
	log.Println(tarballPath)
}
