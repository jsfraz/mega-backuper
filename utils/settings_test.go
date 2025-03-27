package utils

import (
	"jsfraz/mega-backuper/models"
	"testing"
)

func TestValidateSettingsWithDuplicateBackupNames(t *testing.T) {
	lastCopies := 5
	settings := models.BackupSettings{
		Email:    "test@example.com",
		Password: "password",
		Backups: []models.Backup{
			{
				Name:       "backup",
				MegaDir:    "backup/",
				LastCopies: &lastCopies,
				Cron:       "0 12 * * *",
				Type:       "volume",
			},
			{
				Name:       "backup",
				MegaDir:    "backup/",
				LastCopies: &lastCopies,
				Cron:       "0 12 * * *",
			},
		},
	}

	err := ValidateSettings(settings)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestValidateSettings(t *testing.T) {
	lastCopies := 5
	settings := models.BackupSettings{
		Email:    "test@example.com",
		Password: "password",
		Backups: []models.Backup{
			{
				Name:       "backup",
				MegaDir:    "backup/",
				LastCopies: &lastCopies,
				Cron:       "0 12 * * *",
				Type:       "volume",
			},
		},
	}

	err := ValidateSettings(settings)
	if err != nil {
		t.Error(err)
	}
}
