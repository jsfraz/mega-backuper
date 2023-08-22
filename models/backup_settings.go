package models

type BackupSettings struct {
	MegaEmail    string   `json:"megaEmail" validate:"email,required"`
	MegaPassword string   `json:"megaPassword" validate:"required"`
	Backups      []Backup `json:"backups" validate:"required"`
}
