package utils

import (
	"jsfraz/mega-backuper/models"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return data
}

func TestMysqlDSNEscapesCredentials(t *testing.T) {
	backup := models.Backup{
		MysqlUser:     "user@example.com",
		MysqlPassword: `p@ss:w/rd#1 "quoted" \backslash`,
		MysqlHost:     "2001:db8::1",
		MysqlPort:     3307,
		MysqlDb:       "my-db",
	}

	dsn := mysqlDSN(backup)
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("mysqlDSN produced an unparsable DSN %q: %v", dsn, err)
	}
	if cfg.User != backup.MysqlUser {
		t.Errorf("user = %q, want %q", cfg.User, backup.MysqlUser)
	}
	if cfg.Passwd != backup.MysqlPassword {
		t.Errorf("password = %q, want %q", cfg.Passwd, backup.MysqlPassword)
	}
	if cfg.DBName != backup.MysqlDb {
		t.Errorf("database = %q, want %q", cfg.DBName, backup.MysqlDb)
	}
	if want := "[2001:db8::1]:3307"; cfg.Addr != want {
		t.Errorf("address = %q, want %q", cfg.Addr, want)
	}
}

func TestPgDSNEscapesCredentials(t *testing.T) {
	backup := models.Backup{
		PgUser:     "user@example.com",
		PgPassword: `p@ss:w/rd#1 'quoted'`,
		PgHost:     "2001:db8::1",
		PgPort:     5433,
		PgDb:       "my-db",
	}

	dsn := pgDSN(backup)
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("pgDSN produced an unparsable URL %q: %v", dsn, err)
	}
	password, set := parsed.User.Password()
	if !set {
		t.Fatalf("password missing from %q", dsn)
	}
	if password != backup.PgPassword {
		t.Errorf("password = %q, want %q", password, backup.PgPassword)
	}
	if parsed.User.Username() != backup.PgUser {
		t.Errorf("user = %q, want %q", parsed.User.Username(), backup.PgUser)
	}
	if want := "[2001:db8::1]:5433"; parsed.Host != want {
		t.Errorf("host = %q, want %q", parsed.Host, want)
	}
	if strings.TrimPrefix(parsed.Path, "/") != backup.PgDb {
		t.Errorf("database = %q, want %q", parsed.Path, backup.PgDb)
	}
}

func TestQuoteMysqlOptionValue(t *testing.T) {
	quoted, err := quoteMysqlOptionValue(`pa"ss\word`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := `"pa\"ss\\word"`; quoted != want {
		t.Fatalf("quoted = %s, want %s", quoted, want)
	}

	if _, err := quoteMysqlOptionValue("line\nbreak"); err == nil {
		t.Fatal("expected error for a value containing a line break")
	}
}

func TestWriteMysqlDefaultsFile(t *testing.T) {
	backup := models.Backup{
		MysqlUser:     "user",
		MysqlPassword: "secret",
		MysqlHost:     "mysql-example",
		MysqlPort:     3306,
	}
	path := t.TempDir() + "/client.cnf"
	if err := writeMysqlDefaultsFile(path, backup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Fatalf("credential file mode = %o, want 600", mode)
	}

	content := string(readFile(t, path))
	for _, want := range []string{"[client]", `user="user"`, `password="secret"`, `host="mysql-example"`, "port=3306"} {
		if !strings.Contains(content, want) {
			t.Errorf("credential file is missing %q:\n%s", want, content)
		}
	}
}

func TestWritePgPassFileEscapesFields(t *testing.T) {
	backup := models.Backup{
		PgUser:     "user",
		PgPassword: `p:a\ss`,
		PgHost:     "postgres-example",
		PgPort:     5432,
		PgDb:       "postgres",
	}
	path := t.TempDir() + "/pgpass"
	if err := writePgPassFile(path, backup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `postgres-example:5432:postgres:user:p\:a\\ss` + "\n"
	if got := string(readFile(t, path)); got != want {
		t.Fatalf("pgpass = %q, want %q", got, want)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Fatalf("credential file mode = %o, want 600", mode)
	}
}
