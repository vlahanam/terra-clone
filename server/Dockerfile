# Development / Local Dockerfile with hot-reloading via Air
FROM golang:1.26-alpine

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct

# Install tools for development and Air live-reloading
RUN apk add --no-cache git bash curl

# Install Air for hot-reloading
RUN go install github.com/air-verse/air@latest

# Install golang-migrate CLI with postgres support
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Pre-download dependencies to leverage Docker layer caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy application source code
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

