package utils

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"jsfraz/mega-backuper/models"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/t3rm1n4l/go-mega"
)

const (
	// dumpTimeout bounds a single pg_dump/mysqldump invocation.
	dumpTimeout = 2 * time.Hour
	// dbConnectTimeout bounds startup connectivity checks.
	dbConnectTimeout = 15 * time.Second
	// volumeRoot is where volume directories are mounted into the container.
	volumeRoot = "/tmp"
)

var pgDumpVersionRe = regexp.MustCompile(`\(PostgreSQL\) (\d+)`)

// Creates a per-run temporary directory. It is removed by removeWorkDir.
//
//	@param kind Short label used in the directory name.
//	@return string
//	@return error
func newWorkDir(kind string) (string, error) {
	return os.MkdirTemp("", fmt.Sprintf("%s%s-*", workDirPrefix, kind))
}

// Removes a temporary directory, logging instead of returning an error.
//
//	@param path
func removeWorkDir(path string) {
	if err := os.RemoveAll(path); err != nil {
		log.Printf("Failed to remove temporary directory %s: %v", path, err)
	}
}

// Create tarball from source path to target path.
//
//	@param sourcePath
//	@param targetPath
//	@return error
func createTarball(sourcePath string, targetPath string) (err error) {
	// create tarball file
	tarballFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := tarballFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// create a GZIP writer to compress the tarball, then a TAR writer on top
	gzipWriter := gzip.NewWriter(tarballFile)
	tarWriter := tar.NewWriter(gzipWriter)

	if err := writeTarEntries(tarWriter, sourcePath); err != nil {
		return err
	}
	// Closing flushes the archive; errors here mean a truncated tarball.
	if err := tarWriter.Close(); err != nil {
		return err
	}
	return gzipWriter.Close()
}

// Walks sourcePath and writes every entry into the tar writer. Files are
// closed immediately so archiving a large volume does not exhaust file
// descriptors.
//
//	@param tarWriter
//	@param sourcePath
//	@return error
func writeTarEntries(tarWriter *tar.Writer, sourcePath string) error {
	return filepath.Walk(sourcePath, func(filePath string, fileInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// calculate the relative path of the file within the source folder
		relPath, err := filepath.Rel(sourcePath, filePath)
		if err != nil {
			return err
		}
		// the source directory itself is implied by the archive layout
		if relPath == "." {
			return nil
		}

		linkTarget := ""
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			linkTarget, err = os.Readlink(filePath)
			if err != nil {
				return err
			}
		}
		// sockets, FIFOs and device nodes cannot be restored from a tarball
		if !fileInfo.IsDir() && !fileInfo.Mode().IsRegular() && linkTarget == "" {
			log.Printf("Skipping unsupported file type %s (%s)", filePath, fileInfo.Mode().Type())
			return nil
		}

		// create a TAR header based on the file info
		header, err := tar.FileInfoHeader(fileInfo, linkTarget)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		// write the header to the TAR archive
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !fileInfo.Mode().IsRegular() {
			return nil
		}

		// copy the file's content to the TAR archive
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tarWriter, file)
		if closeErr := file.Close(); copyErr == nil {
			copyErr = closeErr
		}
		return copyErr
	})
}

// escapePgpassField escapes a value for use in a .pgpass line.
//
//	@param value
//	@return string
func escapePgpassField(value string) string {
	return strings.NewReplacer(`\`, `\\`, `:`, `\:`).Replace(value)
}

// Writes a 0600 .pgpass file so the database password never appears in the
// process environment.
//
//	@param path
//	@param backup
//	@return error
func writePgPassFile(path string, backup models.Backup) error {
	fields := []string{
		escapePgpassField(backup.PgHost),
		strconv.Itoa(backup.PgPort),
		escapePgpassField(backup.PgDb),
		escapePgpassField(backup.PgUser),
		escapePgpassField(backup.PgPassword),
	}
	return os.WriteFile(path, []byte(strings.Join(fields, ":")+"\n"), 0600)
}

// quoteMysqlOptionValue quotes a value for a MySQL option file.
//
//	@param value
//	@return string
//	@return error
func quoteMysqlOptionValue(value string) (string, error) {
	if strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("value contains a line break")
	}
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return `"` + replacer.Replace(value) + `"`, nil
}

// Writes a 0600 MySQL option file so the database password never appears in
// the process environment.
//
//	@param path
//	@param backup
//	@return error
func writeMysqlDefaultsFile(path string, backup models.Backup) error {
	values := []struct {
		key   string
		value string
	}{
		{"user", backup.MysqlUser},
		{"password", backup.MysqlPassword},
		{"host", backup.MysqlHost},
	}

	var builder strings.Builder
	builder.WriteString("[client]\n")
	for _, entry := range values {
		quoted, err := quoteMysqlOptionValue(entry.value)
		if err != nil {
			return fmt.Errorf("invalid MySQL %s: %w", entry.key, err)
		}
		fmt.Fprintf(&builder, "%s=%s\n", entry.key, quoted)
	}
	fmt.Fprintf(&builder, "port=%d\n", backup.MysqlPort)

	return os.WriteFile(path, []byte(builder.String()), 0600)
}

// mysqlDSN builds a properly escaped MySQL DSN.
//
//	@param backup
//	@return string
func mysqlDSN(backup models.Backup) string {
	cfg := mysql.NewConfig()
	cfg.User = backup.MysqlUser
	cfg.Passwd = backup.MysqlPassword
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(backup.MysqlHost, strconv.Itoa(backup.MysqlPort))
	cfg.DBName = backup.MysqlDb
	cfg.Timeout = dbConnectTimeout
	return cfg.FormatDSN()
}

// pgDSN builds a properly escaped PostgreSQL connection URL.
//
//	@param backup
//	@return string
func pgDSN(backup models.Backup) string {
	connectionURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(backup.PgUser, backup.PgPassword),
		Host:   net.JoinHostPort(backup.PgHost, strconv.Itoa(backup.PgPort)),
		Path:   "/" + backup.PgDb,
	}
	query := connectionURL.Query()
	query.Set("sslmode", "disable")
	query.Set("connect_timeout", strconv.Itoa(int(dbConnectTimeout.Seconds())))
	connectionURL.RawQuery = query.Encode()
	return connectionURL.String()
}

// envWithout returns the process environment without the given variables.
//
//	@param keys
//	@return []string
func envWithout(keys ...string) []string {
	blocked := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		blocked[key] = struct{}{}
	}

	environment := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if _, found := blocked[key]; found {
			continue
		}
		environment = append(environment, entry)
	}
	return environment
}

// Uploads file to Mega and removes the local copy afterwards.
//
//	@param localFilePath
//	@param fileName
//	@param megaDir
//	@return *mega.Node Node where file was uploaded.
//	@return error
func uploadToMegaAndDelete(localFilePath string, fileName string, megaDir string) (*mega.Node, error) {
	defer func() {
		if err := os.Remove(localFilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Failed to remove local file %s: %v", localFilePath, err)
		}
	}()

	// check mega dir and return node to upload to
	uploadNode, err := MegaCheckDir(megaDir)
	if err != nil {
		return nil, err
	}
	if err := MegaUpload(localFilePath, uploadNode, fileName); err != nil {
		return nil, fmt.Errorf("upload of '%s' failed, a partially uploaded file may be left in '%s': %w", fileName, megaDir, err)
	}
	return uploadNode, nil
}

// dumpPostgres runs pg_dump into dumpFilePath.
//
//	@param backup
//	@param dumpFilePath
//	@return error
func dumpPostgres(backup models.Backup, dumpFilePath string) error {
	passFile := filepath.Join(filepath.Dir(dumpFilePath), "pgpass")
	if err := writePgPassFile(passFile, backup); err != nil {
		return fmt.Errorf("failed to write PostgreSQL credential file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dumpTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pg_dump",
		"-Fc",
		"-w",
		"-h", backup.PgHost,
		"-p", strconv.Itoa(backup.PgPort),
		"-U", backup.PgUser,
		"-d", backup.PgDb,
		"-f", dumpFilePath)
	cmd.Env = append(envWithout("PGPASSWORD", "PGPASSFILE"), "PGPASSFILE="+passFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("pg_dump timed out after %s", dumpTimeout)
		}
		return fmt.Errorf("pg_dump failed: %w, output: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// Backup Postgres.
//
//	@param backup
//	@return error
func BackupPostgres(backup models.Backup) error {
	currentTime := time.Now()
	// File name: DB_NAME-UNIX_TIMESTAMP.backup
	dumpFileName := fmt.Sprintf("%s-%d.backup", backup.PgDb, currentTime.Unix())

	workDir, err := newWorkDir("pg")
	if err != nil {
		return err
	}
	defer removeWorkDir(workDir)
	dumpFilePath := filepath.Join(workDir, dumpFileName)

	// Dump database using native pg_dump
	if err := dumpPostgres(backup, dumpFilePath); err != nil {
		return err
	}

	// Upload to mega
	uploadNode, err := uploadToMegaAndDelete(dumpFilePath, dumpFileName, backup.MegaDir)
	if err != nil {
		return err
	}
	// Delete oldest file(s)
	return MegaDeleteFilesByLastCopyCount(backup, uploadNode)
}

// Backup volume to Mega.
//
//	@param backup
//	@return error
func BackupVolume(backup models.Backup) error {
	currentTime := time.Now()
	// make paths
	folderPath := filepath.Join(volumeRoot, backup.Name)
	tarballFileName := fmt.Sprintf("%s-%d.tar.gz", backup.Name, currentTime.Unix())

	workDir, err := newWorkDir("volume")
	if err != nil {
		return err
	}
	defer removeWorkDir(workDir)
	tarballPath := filepath.Join(workDir, tarballFileName)

	// make tarball
	if err := createTarball(folderPath, tarballPath); err != nil {
		return fmt.Errorf("failed to archive volume '%s': %w", folderPath, err)
	}
	// upload to mega
	uploadNode, err := uploadToMegaAndDelete(tarballPath, tarballFileName, backup.MegaDir)
	if err != nil {
		return err
	}
	// delete oldest file(s)
	return MegaDeleteFilesByLastCopyCount(backup, uploadNode)
}

// dumpMysql runs mysqldump into dumpFilePath.
//
//	@param backup
//	@param dumpFilePath
//	@return error
func dumpMysql(backup models.Backup, dumpFilePath string) error {
	defaultsFile := filepath.Join(filepath.Dir(dumpFilePath), "client.cnf")
	if err := writeMysqlDefaultsFile(defaultsFile, backup); err != nil {
		return fmt.Errorf("failed to write MySQL credential file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dumpTimeout)
	defer cancel()

	// --defaults-extra-file has to be the first option
	cmd := exec.CommandContext(ctx, "mysqldump",
		"--defaults-extra-file="+defaultsFile,
		"--single-transaction",
		"--routines",
		"--events",
		"--no-tablespaces",
		"--set-gtid-purged=OFF",
		"--result-file="+dumpFilePath,
		backup.MysqlDb)
	cmd.Env = envWithout("MYSQL_PWD")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("mysqldump timed out after %s", dumpTimeout)
		}
		return fmt.Errorf("mysqldump failed: %w, output: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// Backup Mysql.
//
//	@param backup
//	@return error
func BackupMysql(backup models.Backup) error {
	currentTime := time.Now()
	// File name: DB_NAME-UNIX_TIMESTAMP.sql
	dumpFileName := fmt.Sprintf("%s-%d.sql", backup.MysqlDb, currentTime.Unix())
	tarballFileName := fmt.Sprintf("%s-%d.tar.gz", backup.Name, currentTime.Unix())

	workDir, err := newWorkDir("mysql")
	if err != nil {
		return err
	}
	defer removeWorkDir(workDir)

	// keep the dump in its own directory so the tarball has the same layout as
	// previous releases (a single .sql file at the archive root)
	dumpDir := filepath.Join(workDir, "dump")
	if err := os.Mkdir(dumpDir, 0700); err != nil {
		return err
	}
	dumpFilePath := filepath.Join(dumpDir, dumpFileName)

	// Dump database using native mysqldump
	if err := dumpMysql(backup, dumpFilePath); err != nil {
		return err
	}

	// Make tarball, outside the archived directory
	tarballPath := filepath.Join(workDir, tarballFileName)
	if err := createTarball(dumpDir, tarballPath); err != nil {
		return err
	}

	// Upload to mega
	uploadNode, err := uploadToMegaAndDelete(tarballPath, tarballFileName, backup.MegaDir)
	if err != nil {
		return err
	}
	// Delete oldest file(s)
	return MegaDeleteFilesByLastCopyCount(backup, uploadNode)
}

// CheckConfig validates backup sources and connectivity. Returns an error
// describing the first problem found.
//
//	@return error
func CheckConfig() error {
	// singleton
	s := GetSingleton()
	log.Println("Checking settings...")

	var pgDumpMajor int
	pgDumpMajorLoaded := false

	// cycle trough backups
	for _, backup := range s.Settings.Backups {
		switch backup.Type {

		// Volume
		case models.Volume:
			sourcePath := filepath.Join(volumeRoot, backup.Name)
			info, err := os.Stat(sourcePath)
			if err != nil {
				return fmt.Errorf("could not access directory for job '%s': %w", backup.Name, err)
			}
			if !info.IsDir() {
				return fmt.Errorf("path for job '%s' is not a directory: %s", backup.Name, sourcePath)
			}

		// Postgres
		case models.Postgres:
			if !pgDumpMajorLoaded {
				var err error
				pgDumpMajor, err = getPgDumpMajor()
				if err != nil {
					return fmt.Errorf("failed to determine pg_dump version: %w", err)
				}
				pgDumpMajorLoaded = true
			}
			if err := checkPostgresBackup(backup, pgDumpMajor); err != nil {
				return err
			}

		// Mysql
		case models.Mysql:
			if err := checkMysqlBackup(backup); err != nil {
				return err
			}
		}
	}
	return nil
}

// parsePgDumpMajor extracts the PostgreSQL major version from `pg_dump --version` output.
func parsePgDumpMajor(versionOutput string) (int, error) {
	matches := pgDumpVersionRe.FindStringSubmatch(versionOutput)
	if len(matches) < 2 {
		return 0, fmt.Errorf("could not parse pg_dump version from: %s", strings.TrimSpace(versionOutput))
	}
	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("could not parse pg_dump major version: %w", err)
	}
	return major, nil
}

// pgMajorFromServerVersionNum converts PostgreSQL server_version_num to a major version.
// For PostgreSQL 10+, server_version_num is major*10000 + minor (e.g. 17.11 -> 170011).
func pgMajorFromServerVersionNum(versionNum int) int {
	return versionNum / 10000
}

func getPgDumpMajor() (int, error) {
	output, err := exec.Command("pg_dump", "--version").CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("pg_dump --version failed: %w, output: %s", err, string(output))
	}
	return parsePgDumpMajor(string(output))
}

// checkMysqlBackup verifies that the MySQL server is reachable and accepts
// the configured credentials.
func checkMysqlBackup(backup models.Backup) error {
	db, err := sql.Open("mysql", mysqlDSN(backup))
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection for job '%s': %w", backup.Name, err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dbConnectTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping MySQL for job '%s': %w", backup.Name, err)
	}
	return nil
}

func checkPostgresBackup(backup models.Backup, pgDumpMajor int) error {
	db, err := sql.Open("postgres", pgDSN(backup))
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection for job '%s': %w", backup.Name, err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dbConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL for job '%s': %w", backup.Name, err)
	}

	var versionNumStr string
	if err := db.QueryRowContext(ctx, "SHOW server_version_num").Scan(&versionNumStr); err != nil {
		return fmt.Errorf("failed to read PostgreSQL version for job '%s': %w", backup.Name, err)
	}
	serverVersionNum, err := strconv.Atoi(versionNumStr)
	if err != nil {
		return fmt.Errorf("failed to parse PostgreSQL version for job '%s': %w", backup.Name, err)
	}

	serverMajor := pgMajorFromServerVersionNum(serverVersionNum)
	if !pgDumpSupportsServer(pgDumpMajor, serverMajor) {
		return fmt.Errorf("pg_dump version %d is older than PostgreSQL server %d for job '%s'; pg_dump must be the same major version or newer", pgDumpMajor, serverMajor, backup.Name)
	}
	return nil
}

func pgDumpSupportsServer(dumpMajor, serverMajor int) bool {
	return dumpMajor >= serverMajor
}
