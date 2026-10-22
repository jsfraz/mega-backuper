package utils

import (
	"jsfraz/mega-backuper/models"
	"testing"
)

func TestIsBackupFileFor(t *testing.T) {
	postgres := models.Backup{Name: "pg-job", PgDb: "app", Type: models.Postgres}
	volume := models.Backup{Name: "nginx", Type: models.Volume}
	mysqlBackup := models.Backup{Name: "mariadb", Type: models.Mysql}

	tests := []struct {
		name     string
		backup   models.Backup
		fileName string
		want     bool
	}{
		{name: "postgres archive", backup: postgres, fileName: "app-1700000000.backup", want: true},
		{name: "postgres wrong extension", backup: postgres, fileName: "app-1700000000.tar.gz"},
		{name: "postgres other database", backup: postgres, fileName: "other-1700000000.backup"},
		{name: "postgres missing timestamp", backup: postgres, fileName: "app-.backup"},
		{name: "postgres non numeric timestamp", backup: postgres, fileName: "app-17a.backup"},
		{name: "postgres trailing characters", backup: postgres, fileName: "app-1700000000.backup.bak"},
		{name: "postgres prefixed name", backup: postgres, fileName: "xapp-1700000000.backup"},
		{name: "postgres job name is irrelevant", backup: postgres, fileName: "pg-job-1700000000.backup"},
		{name: "volume archive", backup: volume, fileName: "nginx-1700000000.tar.gz", want: true},
		{name: "volume postgres extension", backup: volume, fileName: "nginx-1700000000.backup"},
		{name: "mysql archive", backup: mysqlBackup, fileName: "mariadb-1700000000.tar.gz", want: true},
		{name: "mysql other job", backup: mysqlBackup, fileName: "nginx-1700000000.tar.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBackupFileFor(tt.backup, tt.fileName); got != tt.want {
				t.Fatalf("isBackupFileFor(%+v, %q) = %v, want %v", tt.backup, tt.fileName, got, tt.want)
			}
		})
	}
}
