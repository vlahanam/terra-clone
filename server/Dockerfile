# [VN] Dockerfile môi trường Dev / Local hỗ trợ live-reload bằng Air
FROM golang:1.26-alpine

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct

# [VN] Cài đặt các công cụ phát triển và hỗ trợ Air
RUN apk add --no-cache git bash curl

# [VN] Cài đặt Air phục vụ hot-reloading
RUN go install github.com/air-verse/air@latest

# [VN] Cài đặt công cụ golang-migrate CLI hỗ trợ PostgreSQL
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0 \
    && ln -sf /go/bin/migrate /usr/local/bin/migrate

# [VN] Tải trước dependencies để tối ưu Docker layer cache
COPY go.mod go.sum* ./
RUN go mod download

# [VN] Sao chép toàn bộ mã nguồn ứng dụng
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

