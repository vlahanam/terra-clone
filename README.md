# Terra Clone - Infrastructure & Deployment Guide

Dự án xây dựng với kiến trúc Microservices / Monorepo bao gồm:

- **Nginx**: Reverse Proxy đóng vai trò Gateway điều hướng traffic tới Backend API và Frontend.
- **Server (Golang)**: Backend API service với tính năng live-reload bằng Air trong môi trường Dev / Local.
- **PostgreSQL 17**: Database chính của hệ thống.
- **Frontend**: Ứng dụng Web (Next.js / React) được điều hướng qua Nginx.

---

## 📌 Lưu ý quan trọng về Môi trường (Environments)

> **Môi trường `dev` và `local` là MỘT.**
>
> - `docker-compose.yml`: Sử dụng cho môi trường **Dev / Local** (mount mã nguồn vào container để code và reload ngay lập tức).
> - `docker-compose.staging.yml`: Sử dụng cho môi trường **Staging** (build image khép kín, tối ưu kiểm thử).
> - `docker-compose.product.yml`: Sử dụng cho môi trường **Production** (build image tối ưu, bảo mật non-root, không public port database ra ngoài host).

---

## 🚀 Cấu trúc thư mục Docker & Nginx

```
terra-clone/
├── .dockerignore                 # Docker ignore cấp root
├── .env.example                  # Template biến môi trường
├── .env                          # Biến môi trường hiện tại
├── docker-compose.yml            # Compose cho Dev / Local
├── docker-compose.staging.yml    # Compose cho Staging
├── docker-compose.product.yml    # Compose cho Production
├── Makefile                      # Các lệnh quản lý Docker thuận tiện
├── nginx/
│   ├── Dockerfile                # Dockerfile Nginx cho Dev / Local
│   ├── Dockerfile.staging        # Dockerfile Nginx cho Staging
│   ├── Dockerfile.product        # Dockerfile Nginx cho Production
│   ├── nginx.conf                # Cấu hình Nginx toàn cục (Gzip, buffer, logging)
│   └── conf.d/
│       ├── default.conf          # Nginx virtual host cho Dev / Local
│       ├── staging.conf          # Nginx virtual host cho Staging
│       └── product.conf          # Nginx virtual host cho Production
├── server/
│   ├── .air.toml                 # Cấu hình Air live-reload cho Golang
│   ├── .dockerignore             # Docker ignore cho server
│   ├── .env.example              # Template biến môi trường server
│   ├── .env                      # Biến môi trường server
│   ├── Dockerfile                # Dockerfile Golang cho Dev / Local (Air)
│   ├── Dockerfile.staging        # Dockerfile Golang cho Staging (Multi-stage)
│   ├── Dockerfile.product        # Dockerfile Golang cho Production (Hardened, non-root)
│   ├── cmd/
│   ├── internal/
│   └── migrations/
└── frontend/
    └── .dockerignore             # Docker ignore cho frontend
```

---

## ⚙️ Cổng dịch vụ (Port Mapping)

Mặc định các cổng được cấu hình trong `.env` để tránh xung đột với các ứng dụng khác:

| Dịch vụ          | Port Container | Port Host (Mặc định) | Ghi chú                                                                  |
| :--------------- | :------------- | :------------------- | :----------------------------------------------------------------------- |
| **Nginx (HTTP)** | `80`           | `8088`               | Truy cập tại `http://localhost:8088` (có thể đổi sang `80` trong `.env`) |
| **Nginx (SSL)**  | `443`          | `8443`               | Truy cập tại `https://localhost:8443`                                    |
| **Server (Go)**  | `8080`         | `8090`               | Port phụ để debug trực tiếp Go server (`http://localhost:8090`)          |
| **PostgreSQL**   | `5432`         | `5432`               | Dùng kết nối DBeaver / TablePlus / psql qua `localhost:5432`             |

---

## 🔀 Quy tắc điều hướng Nginx (Routing Rules)

Nginx được cấu hình với Docker Internal DNS Resolver (`127.0.0.11`) giúp Nginx khởi động mượt mà ngay cả khi một số service upstream chưa sẵn sàng:

1. **Health Check**:
   - `GET /healthz` -> Trả về mã `200 OK` ("healthy") kiểm tra Nginx còn sống hay không.
2. **Backend API**:
   - `GET|POST|... /api/*` -> Reverse proxy tới Go Server (`http://server:8080`).
3. **Frontend Web**:
   - `GET /` và tất cả các route khác -> Reverse proxy tới Frontend (`http://frontend:3000`).
   - Hỗ trợ đầy đủ WebSocket và Upgrade Header cho Hot Module Replacement (HMR).

---

## 🛠️ Hướng dẫn sử dụng Makefile

Để thuận tiện thao tác mà không cần nhớ các câu lệnh dài, dự án đã tích hợp sẵn `Makefile`:

### 1. Môi trường Dev / Local (Mặc định)

```bash
# Khởi chạy toàn bộ hệ thống dev (Nginx, Server, PostgreSQL)
make up
# hoặc
make dev-up

# Xem log thời gian thực
make logs

# Kiểm tra trạng thái các container
make ps

# Dừng môi trường dev
make down

# Build lại container sau khi thay đổi Dockerfile hoặc thư viện
make build
```

### 2. Môi trường Staging & Production

```bash
# Khởi chạy Staging
make staging-up
make staging-logs
make staging-down

# Khởi chạy Production
make prod-up
make prod-logs
make prod-down
```

### 3. Tiện ích (Debug / CLI)

```bash
# Mở terminal sh trong container Go Server
make server-sh

# Mở terminal sh trong container Nginx
make nginx-sh

# Reload Nginx config ngay lập tức (không cần dừng container)
make nginx-reload

# Mở PostgreSQL CLI (psql) tương tác
make psql-cli
# hoặc
make db-cli

# Dọn dẹp toàn bộ container, network và volumes của project
make clean
```
