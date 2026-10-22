# AGENTS.md

Go daemon that dumps PostgreSQL/MySQL databases and tars volumes, uploads them to Mega.nz, and prunes old copies. Module path is `jsfraz/mega-backuper` (not a GitHub path). Entrypoint: `main.go`; settings/Mega/backup logic in `utils/`; structs in `models/`; custom validators in `validators/`.

## Commands

```bash
go build ./...                            # must stay offline-friendly
go test ./...                             # only utils/ has tests
go test ./utils -run TestParsePgDumpMajor -v   # single test
go vet ./...
docker build -t mega-backuper .
```

## Verification limits

- Read `backuper.json` from the **current working directory**, hardcoded: no `--config` flag, no env var. Don't invent config plumbing.
- Running the binary is not a unit test: `utils.CheckConfig()` fatal-exits unless every volume dir `/tmp/<name>` exists, every DB pings, Mega login succeeds, and `pg_dump`/`mysqldump` are on PATH.
- So verify changes with `go build` + `go test`; do not attempt an end-to-end run.
- Tests must stay hermetic: they avoid the `utils` singleton, Mega, and real databases. Don't add tests that need network or credentials.

## Docker

- Runtime base **must stay Debian** (`debian:bookworm-slim`), not Alpine: Alpine only ships the MariaDB client, which fails MySQL 8+ `caching_sha2_password` auth.
- `pg_dump` 18 comes from PGDG (Bookworm's 15 can't dump PG 16+); MySQL client comes from the MySQL APT repo, not Debian.
- Container runs as non-root `app` (UID/GID 1000); keep `--chown=app:app` on copied files.

## Quirks to preserve

- `gocron` is v1 (`gocron.NewScheduler`, `scheduler.Cron(...).Do(...)`). Do not upgrade to v2 API or mix APIs.
- `github.com/t3rm1n4l/go-mega` is a pinned pseudo-version. `destroyOldCopies` is intentionally disabled (two `FIXME` comments reference upstream PR #46); deletions always go to the rubbish bin. Don't re-enable it.
- Global singleton (`utils.GetSingleton()`) carries settings and the Mega client; backup actions receive `models.Backup` by value.
- Backup names must be unique — enforced by a struct-level validator, not struct tags.
- Scheduler and log timestamps are UTC. Output filenames use Unix timestamps.
- Real filename behavior: PostgreSQL dumps are `pgDb-UNIX_TIMESTAMP.backup` (named after the DB, not the job); MySQL/volume archives are `name-UNIX_TIMESTAMP.tar.gz`.

## Repo conventions

- Work on `development` (or feature branches off it). `master` is release-only: pushing to it triggers CI that publishes `ghcr.io/<repo>:latest`.
- Local E2E configs are untracked `backuper_test*.json` (mounted over `/app/backuper.json` by `docker-compose.yaml`). Never commit real Mega/DB credentials or new tracked config variants.
