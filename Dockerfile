# Use the official Golang image as the base image
FROM golang:1.21.0-alpine AS build

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

# Start a new stage using a minimal Alpine image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=build /app/mega-backuper .

# Copy mega-backuper.json
COPY backuper.json /app/backuper.json

# Create temp folder for backups
RUN mkdir temp

# Megacmd
RUN apk add --no-cache megacmd

# Command to run the application
CMD ["./mega-backuper"]