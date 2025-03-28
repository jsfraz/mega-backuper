# mega-backuper

Container for backing up other container's volumes and database dumps to Mega.nz.

## Example usage

> If Mega.nz API returns error 402, login in browser from the same IP address before running the container. (<https://github.com/rclone/rclone/issues/8270#issuecomment-2562047717>)

This example config backups PostgreSQL database from `postgres-example` container every day at 12:00. It keeps last 10 copies in the output directory, older copies are moved to the rubbish bin.

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
        }
    ]
}
```

### Example `docker-compose.yaml`

```yaml
name: mega-backuper-example

services:

  postgres-example:
    image: postgres:alpine
    container_name: postgres-example
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=postgres
    volumes:
      - postgres:/var/lib/postgresql/data
    restart: always

  nginx-example:
    image: nginx:alpine
    container_name: nginx-example
    volumes:
      - nginx:/usr/share/nginx/html
    restart: always

  mega-backuper:
    image: ghcr.io/jsfraz/mega-backuper:latest
    container_name: mega-backuper
    restart: always
    volumes:
      - ./backuper.json:/app/backuper.json
      - nginx:/tmp/nginx

volumes:
  postgres:
  nginx:
```

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
| name             | string | Backup name                                           | true     |
| megaDir          | string | Remote Mega.nz destination directory                  | true     |
| lastCopies       | int    | Number of last copies to keep                         | false    |
| cron             | string | Cron expression for scheduling backup                 | true     |
| type             | string | Backup type (postgres)                                | true     |

<!-- FIXME https://github.com/t3rm1n4l/go-mega/pull/46 -->
<!-- | destroyOldCopies | bool   | Destroy old copies instead moving them to rubbish bin | false    | -->

#### PostgreSQL backup properties

> This project uses [`go-pgdump`](https://github.com/JCoupalK/go-pgdump) to dump PostgreSQL database. It doesn't feature all of `pg_dump` features and only supports dumping table contents, not triggers, views, functions, etc.

| Property   | Type   | Description                                                                              | Required |
|------------|--------|------------------------------------------------------------------------------------------|----------|
| pgUser     | string | PostgreSQL username                                                                      | true     |
| pgPassword | string | PostgreSQL password                                                                      | true     |
| pgDb       | string | PostgreSQL database name                                                                 | true     |
| pgHost     | string | PostgreSQL host (or container name if running in the same network)                       | true     |
| pgPort     | int    | PostgreSQL port                                                                          | true     |

#### Volume backup properties

There are no additional properties for volume backups, however, the `name` property will be used as the directory name mounted to backuper container. Make sure your config looks like this example, where name of the job (`nginx`) is the same as the directory name mounted to backuper container (`/tmp/nginx`).
