package utils

import (
	"archive/tar"
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"jsfraz/mega-backuper/models"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JCoupalK/go-pgdump"
	"github.com/t3rm1n4l/go-mega"
)

// Create tarball and return path to it.
//
//	@param backup
//	@param now
//	@return string Path to tarball.
//	@return string Path to tarball.
//	@return error
func createTarball(backup models.Backup, now time.Time) (string, string, error) {
	// create tarball file
	folderPath := fmt.Sprintf("/tmp/%s", backup.Name)
	// File name: NAME_UNIX_TIMESTAMP.tar.gz
	tarballFileName := fmt.Sprintf("%s_%d.tar.gz", backup.Name, now.Unix())
	tarballPath := fmt.Sprintf("/tmp/%s", tarballFileName)
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
	err = os.Remove(localFilePath)
	if err != nil {
		return nil, err
	}
	return uploadNode, nil
}

// Backup Postgres.
//
//	@param backup
//	@return error
func BackupPostgres(backup models.Backup) error {
	currentTime := time.Now()
	// Init dumper
	dumper := pgdump.NewDumper(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		backup.PgHost, backup.PgPort, backup.PgUser, backup.PgPassword, backup.PgDb), 50)
	// File name: DB_NAME_UNIX_TIMESTAMP.sql
	dumpFilename := fmt.Sprintf("/tmp/%s-%d.sql", backup.PgDb, currentTime.Unix())

	// Dump database
	err := dumper.DumpDatabase(dumpFilename, &pgdump.TableOptions{
		TableSuffix: "",
		TablePrefix: "",
		Schema:      "",
	})
	if err != nil {
		return err
	}

	// Upload to mega
	uploadNode, err := uploadToMegaAndDelete(dumpFilename, strings.Split(dumpFilename, "/tmp/")[1], backup.MegaDir)
	if err != nil {
		return err
	}
	// Delete oldest file(s)
	err = MegaDeleteFilesByLastCopyCount(backup, uploadNode)
	if err != nil {
		return err
	}
	return nil
}

// Backup volume to Mega.
//
//	@param backup
//	@return error
func BackupVolume(backup models.Backup) error {
	currentTime := time.Now()
	// make tarball
	tarballPath, tarballFileName, err := createTarball(backup, currentTime)
	if err != nil {
		return err
	}
	// upload to mega
	uploadNode, err := uploadToMegaAndDelete(tarballPath, tarballFileName, backup.MegaDir)
	if err != nil {
		return err
	}
	// delete oldest file(s)
	err = MegaDeleteFilesByLastCopyCount(backup, uploadNode)
	if err != nil {
		return err
	}
	return nil
}

// Backup Mysql.
//
//	@param backup
//	@return error
func BackupMysql(backup models.Backup) error {
	// TODO mysql dump backup
	return nil
}

// Check if volume directories exists or exit program on fail.
func CheckConfig() {
	// singleton
	s := GetSingleton()
	log.Println("Checking settings...")

	// cycle trough backups
	for _, backup := range s.Settings.Backups {
		switch backup.Type {

		// Volume
		case models.Volume:
			tmp := "/tmp/"
			// check if dir exists
			if _, err := os.Stat(fmt.Sprintf("%s%s", tmp, backup.Name)); os.IsNotExist(err) {
				log.Fatalf("Could not find directory for job '%s': %s%s", backup.Name, tmp, backup.Name)
			}
			break

		// Postgres
		case models.Postgres:
			// ping
			db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
				backup.PgHost, backup.PgPort, backup.PgUser, backup.PgPassword, backup.PgDb))
			if err != nil {
				log.Fatalf("Failed to ping PostgreSQL for job '%s': %v", backup.Name, err)
			}
			db.Close()
			break
		}
	}
}
