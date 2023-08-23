package models

type BackupSettings struct {
	Email    string   `json:"email" validate:"email,required"`
	Password string   `json:"password" validate:"required"`
	Backups  []Backup `json:"backups" validate:"required"`
}
