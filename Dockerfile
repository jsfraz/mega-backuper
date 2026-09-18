# Use the official Golang image as the base image
FROM golang:1.27.1-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go mod and sum files to the working directory
COPY go.mod go.sum ./

# Download and install dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o mega-backuper

# Start a new stage using a minimal Debian image.
# Debian is required because Alpine only ships MariaDB client (incompatible
# with MySQL 8+ caching_sha2_password authentication plugin).
FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

# Install PostgreSQL client from PGDG (Debian Bookworm ships 15, which cannot
# dump PostgreSQL 16+) and the official MySQL community client.
# The MySQL APT repository provides the real `mysqldump` (not the MariaDB
# alias) which supports the caching_sha2_password auth plugin.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        wget \
        gnupg \
        lsb-release \
    && wget -qO - https://www.postgresql.org/media/keys/ACCC4CF8.asc \
        | gpg --dearmor -o /usr/share/keyrings/postgresql-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/postgresql-keyring.gpg] http://apt.postgresql.org/pub/repos/apt bookworm-pgdg main" \
        > /etc/apt/sources.list.d/pgdg.list \
    && wget -qO /tmp/mysql-apt-config.deb \
        https://repo.mysql.com/mysql-apt-config_0.8.39-1_all.deb \
    && echo "mysql-apt-config mysql-apt-config/select-server select mysql-8.4-lts" \
        | debconf-set-selections \
    && dpkg -i /tmp/mysql-apt-config.deb \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        postgresql-client-18 \
        mysql-community-client \
    && apt-get purge -y --auto-remove wget gnupg lsb-release \
    && rm -rf /var/lib/apt/lists/* /tmp/mysql-apt-config.deb \
    && groupadd --gid 1000 app \
    && useradd --uid 1000 --gid 1000 --home-dir /app --no-create-home --shell /usr/sbin/nologin app

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=build --chown=app:app /app/mega-backuper .

# Copy mega-backuper.json
COPY --chown=app:app backuper.json /app/backuper.json

USER app

# Command to run the application
CMD ["./mega-backuper"]
