package models

type BackupType string

const (
	Mysql  BackupType = "mysql"
	Volume BackupType = "volume"
)
