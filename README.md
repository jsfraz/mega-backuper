# mega-backuper

Container for backing up other container's volumes and database dumps to Mega.nz.

## Running in Docker

Configure [backuper.json](backuper.json), [docker-compose.yml](docker-compose.yml), build the image and run the compose project:

```bash
sudo docker build -t jsfraz/mega-backuper:latest .
sudo docker compose up -d
```
