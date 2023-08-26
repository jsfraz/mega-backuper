package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"jsfraz/mega-backuper/models"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/t3rm1n4l/go-mega"
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
	// file name: NAME_RFC3339.tar.gz
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

// Uploads file to Mega and deletes it locally.
//
//	@param localFilePath
//	@param fileName
//	@param megaDir
//	@return *mega.Node Node where file was uploaded.
//	@return error
func uploadToMegaAndDelete(localFilePath string, fileName string, megaDir string) (*mega.Node, error) {
	// check mega dir and return node to upload to
	uploadNode, err := MegaCheckDir(megaDir)
	if err != nil {
		return nil, err
	}
	// upload
	err = MegaUpload(localFilePath, uploadNode, fileName)
	if err != nil {
		return nil, err
	}
	// delete file
	removeErr := os.Remove(localFilePath)
	if removeErr != nil {
		log.Println("Error deleting "+localFilePath+": ", removeErr)
	}
	if err != nil {
		return nil, err
	} else {
		return uploadNode, nil
	}
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
	uploadNode, err := uploadToMegaAndDelete(tarballPath, tarballFileName, backup.MegaDir)
	// delete oldest file(s)
	if backup.LastCopies != 0 {
		deleteErr := MegaDeleteFilesByLastCopyCount(backup, uploadNode)
		if deleteErr != nil {
			log.Println("Error deleting oldest file(s) in "+uploadNode.GetName()+" failed: ", deleteErr)
		}
	}
	return err
}
