package models

type Backup struct {
	Name             string     `json:"name" validate:"alphanum,required"`
	MegaDir          string     `json:"megaDir" validate:"dirpath,required"`
	LastCopies       int        `json:"lastCopies"`
	DestroyOldCopies bool       `json:"destroyOldCopies"`
	Cron             string     `json:"cron" validate:"cron,required"`
	Type             BackupType `json:"type" validate:"oneof=mysql volume,required"`

	// mysql
	MysqlUser     string `json:"mysqlUser" validate:"required_if=Type mysql,omitempty,required"`
	MysqlPassword string `json:"mysqlPassword" validate:"required_if=Type mysql,omitempty,required"`
	MysqlDb       string `json:"mysqlDb" validate:"required_if=Type mysql,omitempty,required"`
	// volume
	// TODO support backuping only selected subdirs of volume
	// Subdirs []string `json:"subdirs"`
}
