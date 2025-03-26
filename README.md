# mega-backuper

Container for backing up other container's volumes and database dumps to Mega.nz.

## Example usage

DISCLAIMER: If Mega.nz API returns error 402, login in browser from the same IP address before running the container. (<https://github.com/meganz/sdk/issues/1433>)

This example config backups PostgreSQL database from `postgres-mega-backuper` container every day at 12:00.

### Example `backuper.json`

```json
{
    "email": "user@example.com",
    "password": "12345678",
    "backups": [
        {
            "name": "postgres",
            "megaDir": "postgres/",
            "cron": "0 12 * * *",
            "type": "postgres",
            "pgUser": "postgres",
            "pgPassword": "postgres",
            "pgDb": "postgres",
            "pgHost": "postgres-example",
            "pgPort": 5432
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

  mega-backuper:
    image: ghcr.io/jsfraz/mega-backuper:latest
    container_name: mega-backuper
    restart: always
    volumes:
      - ./backuper.json:/app/backuper.json

volumes:
  postgres:
```

## Config file properties

### General properties

| Property | Type                | Description                | Required |
|----------|---------------------|----------------------------|----------|
| email    | string              | Your Mega.nz e-mail        | true     |
| password | string              | Your Mega.nz password      | true     |
| backups  | backup object array | Individual backup settings | false    |

### Backup object properties

| Property | Type   | Description                           | Required |
|----------|--------|---------------------------------------|----------|
| name     | string | Backup name                           | true     |
| megaDir  | string | Remote Mega.nz destination directory  | true     |
| cron     | string | Cron expression for scheduling backup | true     |
| type     | string | Backup type (postgres)                | true     |

#### PostgreSQL backup properties

| Property   | Type   | Description                           | Required |
|------------|--------|---------------------------------------|----------|
| pgUser     | string | PostgreSQL username                   | true     |
| pgPassword | string | PostgreSQL password                   | true     |
| pgDb       | string | PostgreSQL database name              | true     |
| pgHost     | string | PostgreSQL host                       | true     |
| pgPort     | int    | PostgreSQL port                       | true     |
