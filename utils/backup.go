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
	"os/exec"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/t3rm1n4l/go-mega"
)

// Create tarball and return path to it.
//
//	@param backup
//	@param now
//	@return string Path to tarball.
//	@return string Path to tarball.
//	@return error
// Create tarball from source path to target path.
//
//	@param sourcePath
//	@param targetPath
//	@return error
func createTarball(sourcePath string, targetPath string) error {
	// create tarball file
	tarballFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer tarballFile.Close()
	// create a GZIP writer to compress the tarball
	gzipWriter := gzip.NewWriter(tarballFile)
	defer gzipWriter.Close()

	// create a TAR writer to write files and headers to the compressed stream
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// walk through the source folder and add its contents to the tarball
	err = filepath.Walk(sourcePath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// create a TAR header based on the file info
		header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return err
		}

		// calculate the relative path of the file within the source folder
		relPath, _ := filepath.Rel(sourcePath, filePath)
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

	return err
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
	// File name: DB_NAME_UNIX_TIMESTAMP.backup
	dumpFileName := fmt.Sprintf("%s-%d.backup", backup.PgDb, currentTime.Unix())
	dumpFilePath := fmt.Sprintf("/tmp/%s", dumpFileName)

	// Dump database using native pg_dump
	cmd := exec.Command("pg_dump", "-Fc", "-h", backup.PgHost, "-p", fmt.Sprintf("%d", backup.PgPort), "-U", backup.PgUser, "-d", backup.PgDb, "-f", dumpFilePath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", backup.PgPassword))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_dump failed: %w, output: %s", err, string(output))
	}

	// Upload to mega
	uploadNode, err := uploadToMegaAndDelete(dumpFilePath, dumpFileName, backup.MegaDir)
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
	// make paths
	folderPath := fmt.Sprintf("/tmp/%s", backup.Name)
	tarballFileName := fmt.Sprintf("%s-%d.tar.gz", backup.Name, currentTime.Unix())
	tarballPath := fmt.Sprintf("/tmp/%s", tarballFileName)

	// make tarball
	err := createTarball(folderPath, tarballPath)
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
	currentTime := time.Now()
	// Make temp folder for tarball source
	tmpFolderPath := fmt.Sprintf("/tmp/%s-%d", backup.Name, currentTime.Unix())
	err := os.MkdirAll(tmpFolderPath, 0755)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpFolderPath)

	// File name: DB_NAME_UNIX_TIMESTAMP.sql
	dumpFileName := fmt.Sprintf("%s-%d.sql", backup.MysqlDb, currentTime.Unix())
	dumpFilePath := fmt.Sprintf("%s/%s", tmpFolderPath, dumpFileName)

	// Dump database using native mysqldump
	cmd := exec.Command("mysqldump", "-h", backup.MysqlHost, "-P", fmt.Sprintf("%d", backup.MysqlPort), "-u", backup.MysqlUser, backup.MysqlDb, "--result-file="+dumpFilePath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", backup.MysqlPassword))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mysqldump failed: %w, output: %s", err, string(output))
	}

	// Make tarball
	tarballFileName := fmt.Sprintf("%s-%d.tar.gz", backup.Name, currentTime.Unix())
	tarballPath := fmt.Sprintf("/tmp/%s", tarballFileName)
	err = createTarball(tmpFolderPath, tarballPath)
	if err != nil {
		return err
	}

	// Upload to mega
	uploadNode, err := uploadToMegaAndDelete(tarballPath, tarballFileName, backup.MegaDir)
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

		// Postgres
		case models.Postgres:
			// ping
			db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
				backup.PgHost, backup.PgPort, backup.PgUser, backup.PgPassword, backup.PgDb))
			if err != nil {
				log.Fatalf("Failed to ping PostgreSQL for job '%s': %v", backup.Name, err)
			}
			db.Close()

		// Mysql
		case models.Mysql:
			// ping
			db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
				backup.MysqlUser, backup.MysqlPassword, backup.MysqlHost, backup.MysqlPort, backup.MysqlDb))
			if err != nil {
				log.Fatalf("Failed to ping MySQL for job '%s': %v", backup.Name, err)
			}
			db.Close()
		}
	}
}
