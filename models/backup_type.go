package models

type BackupType string

const (
	Mysql        BackupType = "mysql"
	ContainerDir BackupType = "containerdir"
)
