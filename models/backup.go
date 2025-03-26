package models

type Backup struct {
	Name    string `json:"name" validate:"alphanum,required"`
	MegaDir string `json:"megaDir" validate:"dirpath,required"`
	// LastCopies       *int       `json:"lastCopies"`
	// DestroyOldCopies *bool      `json:"destroyOldCopies"`
	Cron string     `json:"cron" validate:"cron,required"`
	Type BackupType `json:"type" validate:"oneof=volume postgres mysql,required"`

	// Postgres
	PgUser     string `json:"pgUser" validate:"required_if=Type postgres,omitempty,required"`
	PgPassword string `json:"pgPassword" validate:"required_if=Type postgres,omitempty,required"`
	PgDb       string `json:"pgDb" validate:"required_if=Type postgres,omitempty,required"`
	PgHost     string `json:"pgHost" validate:"required_if=Type postgres,omitempty,required"`
	PgPort     int    `json:"pgPort" validate:"required_if=Type postgres,omitempty,required"`

	// Mysql
	MysqlUser     string `json:"mysqlUser" validate:"required_if=Type mysql,omitempty,required"`
	MysqlPassword string `json:"mysqlPassword" validate:"required_if=Type mysql,omitempty,required"`
	MysqlDb       string `json:"mysqlDb" validate:"required_if=Type mysql,omitempty,required"`

	// Volume
	// TODO Support backuping only selected subdirs of volume
	// Subdirs []string `json:"subdirs"`
}
