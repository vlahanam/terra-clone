# Development / Local Dockerfile with hot-reloading via Air
FROM golang:1.26.7-alpine

WORKDIR /app

# Install tools for development and Air live-reloading
RUN apk add --no-cache git bash curl

# Install Air for hot-reloading
RUN go install github.com/air-verse/air@latest

# Pre-download dependencies to leverage Docker layer caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy application source code
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

