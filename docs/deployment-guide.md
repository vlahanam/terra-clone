# 🚀 Cẩm Nang Triển Khai & Vận Hành Hệ Thống (Deployment Guide)
## Dự án: Terra Clone

---

## 1. 📋 Yêu Cầu Tiền Đề Hệ Thống (Prerequisites)

Trước khi tiến hành cài đặt hoặc triển khai hệ thống **Terra Clone**, đảm bảo máy chủ/môi trường phát triển đáp ứng các yêu cầu sau:

- **Hệ điều hành**: Linux (Ubuntu 22.04+ / Debian 12 / Alpine), macOS hoặc Windows 11 (WSL2).
- **Docker Engine**: Phiên bản `>= 24.0.0`.
- **Docker Compose**: Phiên bản v2 (`docker compose` CLI).
- **GNU Make**: Cài đặt sẵn công cụ `make` để thực thi các lệnh quản trị tự động.
- **Curl / HTTP Client**: Dùng để kiểm tra health check endpoint.

---

## 2. 🌍 Ma Trận Môi Trường (Environments Comparison)

Hệ thống cung cấp cấu hình chuẩn hóa tách biệt cho 3 môi trường phát triển và vận hành:

| Tiêu chí | Dev / Local | Staging | Production |
| :--- | :--- | :--- | :--- |
| **File Compose** | [`docker-compose.yml`](file:///home/wsl-admin/projects/terra-clone/docker-compose.yml) | [`docker-compose.staging.yml`](file:///home/wsl-admin/projects/terra-clone/docker-compose.staging.yml) | [`docker-compose.product.yml`](file:///home/wsl-admin/projects/terra-clone/docker-compose.product.yml) |
| **Dockerfile Server** | [`server/Dockerfile`](file:///home/wsl-admin/projects/terra-clone/server/Dockerfile) | [`server/Dockerfile.staging`](file:///home/wsl-admin/projects/terra-clone/server/Dockerfile.staging) | [`server/Dockerfile.product`](file:///home/wsl-admin/projects/terra-clone/server/Dockerfile.product) |
| **Cơ chế Hot-reload** | Bật qua [Air](file:///home/wsl-admin/projects/terra-clone/server/.air.toml) (mount source `./server:/app`) | Tắt (Build nhị phân) | Tắt (Build nhị phân tối ưu) |
| **Đặc tính Binary** | Debug build thông thường | Multi-stage build | Tối giản, `-trimpath -ldflags="-s -w"` |
| **Bảo mật Container** | User `root` trong container để hot-compile | Non-root `1001:1001` | Hardened Alpine, Non-root `10001:10001` |
| **Mở cổng CSDL** | Mở cổng `5432` ra host để debug tiện lợi | Mở cổng `5432` cho integration test | **Đóng hoàn toàn**, chỉ giao tiếp qua mạng nội bộ |
| **Cấu hình Nginx** | [`conf.d/default.conf`](file:///home/wsl-admin/projects/terra-clone/nginx/conf.d/default.conf) | [`conf.d/staging.conf`](file:///home/wsl-admin/projects/terra-clone/nginx/conf.d/staging.conf) | [`conf.d/product.conf`](file:///home/wsl-admin/projects/terra-clone/nginx/conf.d/product.conf) (Full Security Headers) |
| **Mức độ Logging** | `LevelInfo`, ghi cả Terminal + File | `LevelWarn`, ghi ra File xoay vòng | `LevelWarn`, ghi ra File xoay vòng |

---

## 3. ⚙️ Cấu Hình Biến Môi Trường (Environment Variables)

Hệ thống quản lý biến môi trường theo 2 tầng: Cấp Root (quản lý Docker Compose) và Cấp Server (quản lý nội bộ ứng dụng Go).

### 3.1. Bảng Thông Số Biến Cấp Root ([`.env.example`](file:///home/wsl-admin/projects/terra-clone/.env.example))

| Tên biến | Giá trị mặc định | Mô tả chức năng |
| :--- | :--- | :--- |
| `APP_ENV` | `local` | Định danh môi trường hoạt động (`local`, `staging`, `production`) |
| `SERVER_PORT` | `8080` | Cổng nội bộ mà tiến trình Go Server lắng nghe trong container |
| `SERVER_PORT_EXTERNAL` | `8090` | Cổng host ánh xạ trực tiếp tới Go Server phục vụ kiểm thử debug |
| `DB_HOST` | `postgres` | Host name của dịch vụ CSDL (tên service trong Docker network) |
| `DB_PORT` | `5432` | Cổng nội bộ của PostgreSQL |
| `DB_PORT_EXTERNAL` | `5432` | Cổng host ánh xạ vào PostgreSQL |
| `DB_NAME` | `terra_db` | Tên cơ sở dữ liệu |
| `DB_USER` | `terra_user` | Tên tài khoản quản trị CSDL |
| `DB_PASSWORD` | `terra_secret` | Mật khẩu truy cập CSDL |
| `DB_SSLMODE` | `disable` | Chế độ mã hóa kết nối SSL (`disable`, `require`, `verify-full`) |
| `JWT_SECRET` | *(string 32+ chars)* | Khóa bí mật dùng để ký và xác thực token JWT (HS256) |
| `NGINX_PORT` | `8088` | Cổng HTTP Nginx Gateway mở ra phía ngoài host |
| `NGINX_SSL_PORT` | `8443` | Cổng HTTPS Nginx Gateway |

---

## 4. 🛠️ Quy Trình Triển Khai Chi Tiết (Step-by-Step Deployment)

### 4.1. Bước 1: Khởi Tạo File Môi Trường
Sao chép các tệp mẫu biến môi trường vào vị trí kích hoạt:
```bash
# 1. Cấu hình cấp root cho Docker Compose
cp .env.example .env

# 2. Cấu hình cấp server cho tiến trình Go
cp server/.env.example server/.env
```

> [!NOTE]
> Trên môi trường Staging hoặc Production, cần cập nhật các giá trị `DB_PASSWORD`, `APP_ENV=production`, và `DB_SSLMODE` tương ứng với thông số bảo mật thực tế.

### 4.2. Bước 2: Khởi Chạy Cụm Container Bằng Makefile

#### Khởi chạy môi trường Dev / Local (Mặc định):
```bash
# Build và chạy ngầm toàn bộ dịch vụ (PostgreSQL, Go Server, Nginx)
make up

# Xem danh sách và trạng thái sức khỏe các container
make ps
```

#### Khởi chạy môi trường Staging:
```bash
make staging-build
make staging-up
```

#### Khởi chạy môi trường Production:
```bash
make prod-build
make prod-up
```

### 4.3. Bước 3: Áp Dụng Lược Đồ Dữ Liệu (Run Database Migrations)
Hệ thống sử dụng CLI `golang-migrate` được tích hợp sẵn bên trong container server. Thực thi lệnh sau để nạp toàn bộ cấu trúc bảng:
```bash
make migrate-up DB_URL="postgres://terra_user:terra_secret@postgres:5432/terra_db?sslmode=disable"
```

> [!TIP]
> Bạn có thể export sẵn biến `DB_URL` trong shell để các lệnh sau ngắn gọn hơn:
> ```bash
> export DB_URL="postgres://terra_user:terra_secret@postgres:5432/terra_db?sslmode=disable"
> make migrate-up
> ```

### 4.4. Bước 4: Kiểm Tra Tình Trạng Hoạt Động (Health Check)
Thực thi kiểm tra các điểm cuối:
```bash
# 1. Kiểm tra trạng thái Nginx Gateway
curl -i http://localhost:8088/healthz
# Kết quả mong đợi: HTTP/1.1 200 OK -> healthy

# 2. Kiểm tra API Server thông qua Gateway
curl -i http://localhost:8088/api/ping
# hoặc qua root path:
curl -i http://localhost:8088/ping
# Kết quả mong đợi: HTTP/1.1 200 OK -> {"message":"pong"}

# 3. Kiểm tra API Server trực tiếp qua cổng debug nội bộ
curl -i http://localhost:8090/ping
# hoặc:
curl -i http://localhost:8090/api/ping
# Kết quả mong đợi: HTTP/1.1 200 OK -> {"message":"pong"}
```

---

## 5. 🗃️ Cẩm Nang Quản Lý Database Migrations (`golang-migrate`)

Toàn bộ migration được lưu trữ tại [server/migrations/](file:///home/wsl-admin/projects/terra-clone/server/migrations).

```bash
# 1. Tạo cặp file migration mới (tự động tạo tệp .up.sql và .down.sql đánh số tuần tự)
make migration name=add_refresh_tokens_table

# 2. Áp dụng tất cả các file migration mới nhất
make migrate-up

# 3. Hoàn tác (rollback) duy nhất 1 bước migration gần nhất
make migrate-down

# 4. Xóa sạch toàn bộ dữ liệu (Drop tables) và chạy lại toàn bộ từ version 000001
# CẢNH BÁO: Chỉ sử dụng trên môi trường Local Dev!
make migrate-reset
```

---

## 6. 🩺 Quản Trị, Giám Sát & Khắc Phục Sự Cố (Troubleshooting)

### 6.1. Theo Dõi Logs Theo Thời Gian Thực
```bash
# Xem log tổng hợp của tất cả container Dev
make logs

# Xem log riêng lẻ của Go Server
docker compose logs -f server

# Xem log riêng lẻ của Nginx
docker compose logs -f nginx
```

### 6.2. Kiểm Tra Nhật Ký File Trong Container
Các file log tự động được ghi nhận tại thư mục `server/logs/`:
- `server/logs/access.log`: Ghi nhận toàn bộ HTTP request (IP, Latency, Status, User-Agent).
- `server/logs/app.log`: Ghi nhận structured log JSON của ứng dụng.

### 6.3. Truy Cập Terminal Container
```bash
# Truy cập container Go Server
make server-sh

# Truy cập container Nginx
make nginx-sh

# Nạp lại cấu hình Nginx tức thời mà không dừng dịch vụ
make nginx-reload

# Mở giao diện dòng lệnh PostgreSQL CLI (psql)
make psql-cli
```

### 6.4. Các Sự Cố Phổ Biến & Hướng Xử Lý

1. **Lỗi "Database is dirty" khi chạy migration**:
   - *Nguyên nhân*: Một script migration trước đó bị lỗi cú pháp giữa chừng khiến `schema_migrations` bị khóa cờ dirty.
   - *Xử lý*: Truy cập container qua `make server-sh` và chạy lệnh ép phiên bản:
     `migrate -path migrations -database "$DB_URL" force <version_hợp_lệ>`
2. **Lỗi đụng độ cổng (Port already allocated)**:
   - *Nguyên nhân*: Cổng `8088` hoặc `5432` đã bị một dịch vụ khác trên máy host chiếm dụng.
   - *Xử lý*: Mở tệp `.env` cấp root, thay đổi `NGINX_PORT` thành cổng khác (ví dụ: `8099`) hoặc `DB_PORT_EXTERNAL=5433`, sau đó chạy `make dev-restart`.
3. **Lỗi quyền ghi thư mục logs (`permission denied`)**:
   - *Nguyên nhân*: Tiến trình non-root trong container không có quyền tạo file trong thư mục `server/logs`.
   - *Xử lý*: Chạy trên host: `chmod -R 777 server/logs`.
