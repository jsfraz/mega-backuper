package models

type Backup struct {
	Name       string     `json:"name" validate:"alphanum,required"`
	MegaDir    string     `json:"megaDir" validate:"dirpath,required"`
	LastCopies int        `json:"lastCopies" validate:"gt=0,required"`
	Cron       string     `json:"cron" validate:"cron,required"`
	Type       BackupType `json:"type" validate:"oneof=mysql volume,required"`

	// mysql (validate if Type == mysql)
	MysqlUser     string `json:"mysqlUser" validate:"required_if=Type mysql,omitempty,required"`
	MysqlPassword string `json:"mysqlPassword" validate:"required_if=Type mysql,omitempty,required"`
	MysqlDb       string `json:"mysqlDb" validate:"required_if=Type mysql,omitempty,required"`
	// volume (validate if Type == volume)
	Volume  string   `json:"volume" validate:"required_if=Type volume,omitempty"`
	Subdirs []string `json:"subdirs"`
}
