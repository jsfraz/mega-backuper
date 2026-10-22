package utils

import (
	"jsfraz/mega-backuper/models"
	"testing"
)

func validBackup(name string) models.Backup {
	lastCopies := 5
	return models.Backup{
		Name:       name,
		MegaDir:    "backup/",
		LastCopies: &lastCopies,
		Cron:       "0 12 * * *",
		Type:       "volume",
	}
}

func TestValidateSettingsWithDuplicateBackupNames(t *testing.T) {
	settings := models.BackupSettings{
		Email:    "test@example.com",
		Password: "password",
		Backups: []models.Backup{
			validBackup("backup"),
			validBackup("backup"),
		},
	}

	err := ValidateSettings(settings)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestValidateSettings(t *testing.T) {
	settings := models.BackupSettings{
		Email:    "test@example.com",
		Password: "password",
		Backups:  []models.Backup{validBackup("backup")},
	}

	err := ValidateSettings(settings)
	if err != nil {
		t.Error(err)
	}
}

func TestValidateSettingsRejectsUnsafeFileNames(t *testing.T) {
	unsafeNames := []string{
		"../evil",
		"a/b",
		"-leading-dash",
		".hidden",
		"backup with spaces",
		`backup\windows`,
	}

	for _, name := range unsafeNames {
		t.Run(name, func(t *testing.T) {
			settings := models.BackupSettings{
				Email:    "test@example.com",
				Password: "password",
				Backups:  []models.Backup{validBackup(name)},
			}
			if err := ValidateSettings(settings); err == nil {
				t.Fatalf("expected validation error for backup name %q", name)
			}
		})
	}
}

func TestValidateSettingsRejectsUnsafeDatabaseName(t *testing.T) {
	backup := models.Backup{
		Name:          "postgres",
		MegaDir:       "postgres/",
		Cron:          "0 12 * * *",
		Type:          models.Postgres,
		PgUser:        "postgres",
		PgPassword:    "postgres",
		PgDb:          "../escape",
		PgHost:        "postgres-example",
		PgPort:        5432,
		MysqlUser:     "user",
		MysqlPassword: "password",
		MysqlDb:       "db",
		MysqlHost:     "mysql-example",
		MysqlPort:     3306,
	}
	settings := models.BackupSettings{
		Email:    "test@example.com",
		Password: "password",
		Backups:  []models.Backup{backup},
	}

	if err := ValidateSettings(settings); err == nil {
		t.Fatal("expected validation error for an unsafe pgDb")
	}
}

func TestValidateSettingsAcceptsSingleCharacterName(t *testing.T) {
	settings := models.BackupSettings{
		Email:    "test@example.com",
		Password: "password",
		Backups:  []models.Backup{validBackup("a")},
	}

	if err := ValidateSettings(settings); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
