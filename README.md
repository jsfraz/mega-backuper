# mega-backuper

Container for backing up other container's volumes and database dumps to Mega.nz.

## Example usage

> If Mega.nz API returns error 402, login in browser from the same IP address before running the container. (<https://github.com/rclone/rclone/issues/8270#issuecomment-2562047717>)

This example config backups PostgreSQL database from `postgres-example` container every day at 12:00, and MariaDB database from `mariadb-example` every day at 14:00. It keeps last 10 copies in the output directory, older copies are moved to the rubbish bin.

The `nginx` backup will backup the `nginx-example` container's html directory every day at 10:00.

### Example `backuper.json`

```json
{
    "email": "user@example.com",
    "password": "12345678",
    "backups": [
        {
            "name": "postgres",
            "megaDir": "postgres/",
            "lastCopies": 10,
            "cron": "0 12 * * *",
            "type": "postgres",
            "pgUser": "postgres",
            "pgPassword": "postgres",
            "pgDb": "postgres",
            "pgHost": "postgres-example",
            "pgPort": 5432
        },
        {
            "name": "nginx",
            "megaDir": "nginx/",
            "lastCopies": 5,
            "cron": "0 10 * * *",
            "type": "volume"
        },
        {
            "name": "mariadb",
            "megaDir": "mariadb/",
            "lastCopies": 10,
            "cron": "0 14 * * *",
            "type": "mysql",
            "mysqlUser": "mariadb",
            "mysqlPassword": "mariadb",
            "mysqlDb": "mariadb",
            "mysqlHost": "mariadb-example",
            "mysqlPort": 3306
        }
    ]
}
```

### Example `docker-compose.yaml`

<https://github.com/jsfraz/mega-backuper/blob/master/docker-compose.yaml>

> The image does not ship a `backuper.json`; mount your config to `/app/backuper.json`. The container exits with a clear error when the file is missing.

## Export formats

The output files are named `UNIX_TIMESTAMP`-suffixed: PostgreSQL dumps use the database name (`pgDb-UNIX_TIMESTAMP.backup`), MySQL and volume archives use the job name (`name-UNIX_TIMESTAMP.tar.gz`).

| Backup Type | Extension | Internal Format                                                         |
|-------------|-----------|-------------------------------------------------------------------------|
| `postgres`  | `.backup` | Custom PostgreSQL binary format (`pg_dump -Fc`). Compressed internally. |
| `mysql`     | `.tar.gz` | Gunzipped tarball containing `.sql` dump.                               |
| `volume`    | `.tar.gz` | Gunzipped tarball containing volume contents.                           |

Old copies are pruned from `megaDir` after a successful upload when `lastCopies` is set. Only files produced by the same job are considered (PostgreSQL: `pgDb-UNIX_TIMESTAMP.backup`, MySQL/volume: `name-UNIX_TIMESTAMP.tar.gz`), so several jobs may share one `megaDir` without deleting each other's backups. Directories are never touched.


## Config file properties

### General properties

| Property | Type                | Description                | Required |
|----------|---------------------|----------------------------|----------|
| email    | string              | Your Mega.nz e-mail        | true     |
| password | string              | Your Mega.nz password      | true     |
| backups  | backup object array | Individual backup settings | true     |

### Backup object properties

| Property         | Type   | Description                                           | Required |
|------------------|--------|-------------------------------------------------------|----------|
| name             | string | Backup name (letters, digits, `.`, `_`, `-`)          | true     |
| megaDir          | string | Remote Mega.nz destination directory                  | true     |
| lastCopies       | int    | Number of this job's archives to keep                 | false    |
| cron             | string | Cron expression (5 fields, or 6 with leading seconds) | true     |
| runOnStart       | bool   | Run this backup once when the program starts          | false    |
| type             | string | Backup type (postgres, mysql, volume)                 | true     |

`cron` accepts standard 5-field expressions, 6-field expressions with a leading seconds field, and descriptors such as `@daily` or `@every 1h`. Expressions with a year field (7 fields) are rejected at startup. Invalid expressions and invalid config values stop the container immediately instead of silently skipping a job.

Running the same job twice in parallel is prevented; different jobs may run at the same time.

<!-- FIXME https://github.com/t3rm1n4l/go-mega/pull/46 -->
<!-- | destroyOldCopies | bool   | Destroy old copies instead moving them to rubbish bin | false    | -->

#### PostgreSQL backup properties

> This project uses native `pg_dump` to dump PostgreSQL database, so it supports all of `pg_dump` features like triggers, views, functions, etc. The image ships `pg_dump` 18; the PostgreSQL server major version must be 18 or older (a newer client can dump older servers). The password is passed through a temporary `0600` `.pgpass` file, never through the environment. A dump that takes longer than 2 hours is aborted.

| Property   | Type   | Description                                                                              | Required |
|------------|--------|------------------------------------------------------------------------------------------|----------|
| pgUser     | string | PostgreSQL username                                                                      | true     |
| pgPassword | string | PostgreSQL password                                                                      | true     |
| pgDb       | string | PostgreSQL database name                                                                 | true     |
| pgHost     | string | PostgreSQL host (or container name if running in the same network)                       | true     |
| pgPort     | int    | PostgreSQL port                                                                          | true     |

#### MySQL/MariaDB backup properties

> This project uses native `mysqldump` to dump MySQL/MariaDB database. The dump is taken in a single transaction (`--single-transaction`) together with routines and events, so InnoDB tables are consistent; non-transactional tables (MyISAM/Aria) may still change during the dump. The password is passed through a temporary `0600` option file, never through the environment. A dump that takes longer than 2 hours is aborted.

| Property      | Type   | Description                                                        | Required |
|---------------|--------|--------------------------------------------------------------------|----------|
| mysqlUser     | string | MySQL username                                                     | true     |
| mysqlPassword | string | MySQL password                                                     | true     |
| mysqlDb       | string | MySQL database name                                                | true     |
| mysqlHost     | string | MySQL host (or container name if running in the same network)      | true     |
| mysqlPort     | int    | MySQL port                                                         | true     |

#### Volume backup properties

There are no additional properties for volume backups, however, the `name` property will be used as the directory name mounted to backuper container. Make sure your config looks like this example, where name of the job (`nginx`) is the same as the directory name mounted to backuper container (`/tmp/nginx`). Symlinks are stored as symlinks (their target is preserved, not dereferenced); sockets, FIFOs and device nodes are skipped with a warning.
