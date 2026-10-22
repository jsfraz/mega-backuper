package models

type Backup struct {
	Name       string `json:"name" validate:"required,safesegment"`
	MegaDir    string `json:"megaDir" validate:"dirpath,required"`
	LastCopies *int   `json:"lastCopies" validate:"omitempty,gt=0"`
	// FIXME https://github.com/t3rm1n4l/go-mega/pull/46
	// DestroyOldCopies bool       `json:"destroyOldCopies" validate:"required_with=LastCopies"`
	Cron       string     `json:"cron" validate:"cron,required"`
	RunOnStart bool       `json:"runOnStart"`
	Type       BackupType `json:"type" validate:"oneof=postgres volume mysql,required"`

	// Postgres
	PgUser     string `json:"pgUser" validate:"required_if=Type postgres,omitempty"`
	PgPassword string `json:"pgPassword" validate:"required_if=Type postgres,omitempty"`
	PgDb       string `json:"pgDb" validate:"required_if=Type postgres,omitempty,safesegment"`
	PgHost     string `json:"pgHost" validate:"required_if=Type postgres,omitempty"`
	PgPort     int    `json:"pgPort" validate:"required_if=Type postgres,omitempty"`

	// Mysql
	MysqlUser     string `json:"mysqlUser" validate:"required_if=Type mysql,omitempty"`
	MysqlPassword string `json:"mysqlPassword" validate:"required_if=Type mysql,omitempty"`
	MysqlDb       string `json:"mysqlDb" validate:"required_if=Type mysql,omitempty,safesegment"`
	MysqlHost     string `json:"mysqlHost" validate:"required_if=Type mysql,omitempty"`
	MysqlPort     int    `json:"mysqlPort" validate:"required_if=Type mysql,omitempty"`
}
