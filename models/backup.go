package models

type Backup struct {
	Name       string `json:"name" validate:"required"`
	MegaDir    string `json:"megaDir" validate:"dirpath,required"`
	LastCopies *int   `json:"lastCopies" validate:"omitempty,gt=0"`
	// FIXME https://github.com/t3rm1n4l/go-mega/pull/46
	// DestroyOldCopies bool       `json:"destroyOldCopies" validate:"required_with=LastCopies"`
	Cron string     `json:"cron" validate:"cron,required"`
	Type BackupType `json:"type" validate:"oneof=postgres volume,required"`

	// Postgres
	PgUser     string `json:"pgUser" validate:"required_if=Type postgres,omitempty,required"`
	PgPassword string `json:"pgPassword" validate:"required_if=Type postgres,omitempty,required"`
	PgDb       string `json:"pgDb" validate:"required_if=Type postgres,omitempty,required"`
	PgHost     string `json:"pgHost" validate:"required_if=Type postgres,omitempty,required"`
	PgPort     int    `json:"pgPort" validate:"required_if=Type postgres,omitempty,required"`

	// TODO mysql
	// Mysql
	// MysqlUser     string `json:"mysqlUser" validate:"required_if=Type mysql,omitempty,required"`
	// MysqlPassword string `json:"mysqlPassword" validate:"required_if=Type mysql,omitempty,required"`
	// MysqlDb       string `json:"mysqlDb" validate:"required_if=Type mysql,omitempty,required"`
}
