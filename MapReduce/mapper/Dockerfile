# syntax=docker/dockerfile:1

# Use official Go base image
FROM golang:1.25.1

# Set working directory inside the container
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY *.go ./

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o /service

# Document the port the app listens on
EXPOSE 8080

# Start the app when the container runs
CMD ["/service"]