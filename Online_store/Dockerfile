# syntax=docker/dockerfile:1

# Use official Go base image (includes Go compiler and tooling)
FROM golang:1.25.1

# Set working directory inside the container
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY *.go ./

# Build the Go app (binary will be placed in /myapp)
RUN CGO_ENABLED=0 GOOS=linux go build -o /main

# Document the port the app listens on
EXPOSE 8080

# Start the app when the container runs
CMD ["/main"]
