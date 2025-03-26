package models

type BackupType string

const (
	// TODO volumes and mysql
	// Volume   BackupType = "volume"
	Postgres BackupType = "postgres"
	// Mysql    BackupType = "mysql"
)
