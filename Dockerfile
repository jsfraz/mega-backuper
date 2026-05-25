# Use the official Golang image as the base image
FROM golang:1.26.3-alpine AS build

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

# Install postgresql-client and the official MySQL community client.
# The MySQL APT repository provides the real `mysqldump` (not the MariaDB
# alias) which supports the caching_sha2_password auth plugin.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        wget \
        gnupg \
        lsb-release \
        postgresql-client \
    && wget -qO /tmp/mysql-apt-config.deb \
        https://repo.mysql.com/mysql-apt-config_0.8.39-1_all.deb \
    && echo "mysql-apt-config mysql-apt-config/select-server select mysql-8.4-lts" \
        | debconf-set-selections \
    && dpkg -i /tmp/mysql-apt-config.deb \
    && apt-get update \
    && apt-get install -y --no-install-recommends mysql-community-client \
    && apt-get purge -y --auto-remove wget gnupg lsb-release \
    && rm -rf /var/lib/apt/lists/* /tmp/mysql-apt-config.deb

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=build /app/mega-backuper .

# Copy mega-backuper.json
COPY backuper.json /app/backuper.json

# Command to run the application
CMD ["./mega-backuper"]
