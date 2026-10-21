# Terra Clone - Infrastructure & Backend Architecture Guide

Dự án **Terra Clone** được thiết kế và xây dựng theo mô hình Monorepo / Microservices hiện đại, tối ưu cho quy trình phát triển từ Local/Dev đến Staging và Production.

---

## 📑 Mục lục

1. [Tổng quan Kiến trúc & Tech Stack](#-tổng-quan-kiến-trúc--tech-stack)
2. [Lưu ý quan trọng về Môi trường](#-lưu-ý-quan-trọng-về-môi-trường-environments)
3. [Cấu trúc Thư mục Dự án](#-cấu-trúc-thư-mục-dự-án)
4. [Cổng Dịch vụ & Quy tắc Điều hướng](#-cổng-dịch-vụ--quy-tắc-điều-hướng)
5. [Cơ sở Dữ liệu & Migrations](#-cơ-sở-dữ-liệu--migrations-golang-migrate)
6. [Kiến trúc Backend & Tiêu chuẩn Mã nguồn](#-kiến-trúc-backend--tiêu-chuẩn-mã-nguồn)
7. [Hướng dẫn Bắt đầu Nhanh (Quick Start)](#-hướng-dẫn-bắt-đầu-nhanh-quick-start)
8. [Danh mục Lệnh Makefile Toàn diện](#-danh-mục-lệnh-makefile-toàn-diện)
9. [Quy Trình Phát Triển & Chuẩn Mực Agent (AGENTS.md)](#-quy-trình-phát-triển--chuẩn-mực-agent-agentsmd)

---

## 🏛️ Tổng quan Kiến trúc & Tech Stack

- **Nginx (Reverse Proxy & API Gateway)**:
  - Cổng giao tiếp duy nhất đón nhận request từ client và điều hướng lưu lượng.
  - Tích hợp Docker Internal DNS Resolver (`127.0.0.11`) để đảm bảo không bị lỗi khởi động khi các service nội bộ chưa sẵn sàng.
  - Hỗ trợ đầy đủ WebSocket và Header Upgrade cho Hot Module Replacement (HMR).
- **Backend API Server (Golang 1.26+)**:
  - Web framework: [Gin Gonic](https://github.com/gin-gonic/gin).
  - ORM: [GORM](https://gorm.io/) với PostgreSQL driver kết hợp Connection Pool.
  - Quản lý cấu hình: `godotenv`.
  - Logging: `log/slog` (JSON format) kết hợp `lumberjack/v3` xoay vòng và nén file log tự động.
  - Live Reload (Dev): Tích hợp [Air](https://github.com/air-verse/air) theo dõi và tự động reload mã nguồn.
  - Kiến trúc phân tầng (Clean / Layered Architecture): Controller - Service - Repository.
- **Database (PostgreSQL 18 Alpine)**:
  - Khởi tạo mặc định với phiên bản PostgreSQL 18 trên nền Alpine Linux nhẹ và bảo mật.
  - Thiết kế bảng sử dụng khóa chính `UUIDv7`, hỗ trợ tìm kiếm Social Login, tối ưu hóa JSONB với `GIN Index`.
- **Database Migration ([golang-migrate/v4](https://github.com/golang-migrate/migrate))**:
  - Tích hợp trực tiếp CLI `migrate` vào container server ở môi trường Dev để quản lý lược đồ CSDL theo version (`up`/`down`).
- **Frontend (Web Application)**:
  - Ứng dụng Next.js / React được Nginx điều hướng và ánh xạ tự động.

---

## 📌 Lưu ý quan trọng về Môi trường (Environments)

> **Môi trường `dev` và `local` là MỘT.**
>
> - `docker-compose.yml`: Dành riêng cho môi trường **Dev / Local** (mount trực tiếp mã nguồn vào container để code và nhận live-reload ngay lập tức bằng Air).
> - `docker-compose.staging.yml`: Dành cho môi trường **Staging** (sử dụng Dockerfile multi-stage, chạy non-root user 1001:1001, phục vụ kiểm thử tích hợp).
> - `docker-compose.product.yml`: Dành cho môi trường **Production** (sử dụng Dockerfile hardened, build nhị phân `-trimpath` & `-ldflags="-s -w"`, user bảo mật 10001:10001, không mở port database trực tiếp ra ngoài host).

---

## 🚀 Cấu trúc Thư mục Dự án

```
terra-clone/
├── .dockerignore                 # Docker ignore cấp root
├── .env.example                  # Template biến môi trường toàn dự án
├── .env                          # Biến môi trường hiện tại của Docker Compose
├── AGENTS.md                     # Quy chuẩn & quy trình phát triển bắt buộc cho AI Agent
├── docker-compose.yml            # Compose cho Dev / Local (Air live-reload)
├── docker-compose.staging.yml    # Compose cho Staging (Multi-stage test)
├── docker-compose.product.yml    # Compose cho Production (Hardened, non-root)
├── Makefile                      # Trình quản lý toàn bộ tác vụ CLI
├── plans/                        # Kế hoạch chi tiết từng chức năng trước khi code
├── docs/                         # Toàn bộ tài liệu kiến trúc, PDR, roadmap & guidelines
├── nginx/
│   ├── Dockerfile                # Dockerfile Nginx cho Dev / Local
│   ├── Dockerfile.staging        # Dockerfile Nginx cho Staging
│   ├── Dockerfile.product        # Dockerfile Nginx cho Production
│   ├── nginx.conf                # Cấu hình Nginx toàn cục (Gzip, logging, client buffer)
│   └── conf.d/
│       ├── default.conf          # Nginx virtual host cho Dev / Local
│       ├── staging.conf          # Nginx virtual host cho Staging
│       └── product.conf          # Nginx virtual host cho Production
├── server/
│   ├── .air.toml                 # Cấu hình Air hot-reloading cho Go
│   ├── .dockerignore             # Docker ignore riêng cho server
│   ├── .env.example              # Template cấu hình môi trường cho server
│   ├── .env                      # Cấu hình môi trường nội bộ của Go server
│   ├── Dockerfile                # Dev Dockerfile (Cài Air + golang-migrate CLI)
│   ├── Dockerfile.staging        # Multi-stage Dockerfile cho Staging
│   ├── Dockerfile.product        # Minimal hardened Dockerfile cho Production
│   ├── go.mod                    # Quản lý Golang module dependencies
│   ├── go.sum                    # Checksum Golang module
│   ├── cmd/
│   │   └── main.go               # Điểm khởi chạy chương trình (Entry point)
│   ├── internal/                 # Các gói nội bộ cốt lõi của ứng dụng
│   │   ├── app.go                # Định nghĩa AppConfig cấu trúc
│   │   ├── config.go             # Đọc cấu hình từ .env & biến môi trường hệ thống
│   │   ├── logger.go             # Khởi tạo slog JSON & lumberjack log-roller
│   │   ├── postgresql.go         # Kết nối GORM và kiểm tra Ping Connection Pool
│   │   ├── router.go             # Cấu hình Engine Gin và khai báo các route cơ sở
│   │   └── run.go                # Khởi động ứng dụng và quản lý lifecycle cleanup
│   ├── common/                   # Cấu trúc dữ liệu và chuẩn phản hồi dùng chung
│   │   ├── error_response.go     # Chuẩn hóa lỗi API, Error Carriers, StackTrace
│   │   ├── id.go                 # Hàm sinh public_id ngẫu nhiên an toàn
│   │   ├── jwt.go                # Phát sinh và xác thực token JWT (Access/Refresh)
│   │   ├── jwt_test.go           # Kiểm thử đơn vị JWT và sinh public_id
│   │   ├── pagination.go         # Cấu trúc phân trang Cursor (Paging)
│   │   └── success_response.go   # Chuẩn hóa phản hồi thành công (Data, Paging, Extra)
│   ├── controllers/              # Tầng tiếp nhận HTTP Request & trả về Response
│   │   ├── auth_controller.go    # Controller Đăng ký, Đăng nhập, Refresh, Logout-All, Me
│   │   └── user_controller.go    # Controller CRUD User, phân trang, lọc và tìm kiếm
│   ├── dto/                      # Khối Data Transfer Objects (Request/Response Binding)
│   │   ├── auth_dto.go           # DTO Register, Login, Refresh, AuthResponse
│   │   └── user_dto.go           # DTO CreateUser, UpdateUser, ListUsers, UserResponse
│   ├── models/                   # Entity models tương tác cơ sở dữ liệu qua GORM
│   │   ├── credential.go         # Entity UserCredentials (hash, token_version, lock)
│   │   ├── profile.go            # Entity UserProfile (hồ sơ, preferences JSONMap)
│   │   └── user.go               # Entity User (id UUIDv7, public_id, email, status)
│   ├── services/                 # Tầng xử lý nghiệp vụ kinh doanh (Business Logic)
│   │   ├── auth_service.go       # Logic Auth, transaction 3 bảng, chống brute-force
│   │   └── user_service.go       # Logic CRUD User, cursor pagination, soft-delete
│   ├── repositories/             # Tầng tương tác truy vấn dữ liệu (Data Access)
│   │   ├── credential_repository.go # Truy vấn credentials, đếm sai pass, tăng token version
│   │   ├── profile_repository.go # Truy vấn hồ sơ người dùng, upsert JSONB
│   │   └── user_repository.go    # Truy vấn users, tìm kiếm, phân trang cursor UUIDv7
│   ├── middleware/               # Tầng middleware xử lý HTTP
│   │   ├── auth.go               # Middleware RequireAuth kiểm tra JWT & token_version
│   │   ├── logger.go             # Custom Gin Access Logger format chi tiết
│   │   └── request_id.go         # Middleware sinh và gán X-Request-ID cho mỗi request
│   ├── migrations/               # Các tệp SQL migration có đánh số thứ tự
│   │   ├── 000001_users_table.up.sql / .down.sql
│   │   ├── 000002_user_credentials_table.up.sql / .down.sql
│   │   └── 000003_user_profiles_table.up.sql / .down.sql
│   └── logs/
│       ├── access.log            # Access log chi tiết theo thời gian thực
│       └── app.log               # Structured log JSON của ứng dụng (slog)
└── frontend/
    └── .dockerignore             # Docker ignore cho thư mục frontend
```

---

## ⚙️ Cổng Dịch vụ & Quy tắc Điều hướng

### 1. Bảng phân bổ Cổng (Port Mapping)

Mặc định các cổng host được định nghĩa trong file `.env` cấp root nhằm tránh đụng độ với các dịch vụ đang chạy sẵn trên máy:

| Dịch vụ          | Cổng Container | Cổng Host (Mặc định) | Mô tả & Cách truy cập                                           |
| :--------------- | :------------- | :------------------- | :-------------------------------------------------------------- |
| **Nginx (HTTP)** | `80`           | `8088`               | Điểm vào chính: `http://localhost:8088` (có thể đổi thành `80`) |
| **Nginx (SSL)**  | `443`          | `8443`               | Truy cập an toàn: `https://localhost:8443`                      |
| **Server (Go)**  | `8080`         | `8090`               | Cổng phụ debug trực tiếp backend: `http://localhost:8090`       |
| **PostgreSQL**   | `5432`         | `5432`               | Kết nối DBeaver / TablePlus / psql: `localhost:5432`            |

### 2. Quy tắc Điều hướng Nginx (Routing Rules)

- `GET /healthz`: Trả về trạng thái `200 OK` ("healthy\n") để kiểm tra tình trạng sống còn của Nginx Proxy.
- `GET /ping`: Chuyển tiếp tới Go Server (`http://server:8080/ping`) trả về `{"message":"pong"}` phục vụ kiểm tra liveness.
- `ALL /api/*`: Reverse proxy tới Go Server (`http://server:8080`), tự động chuyển tiếp các header `Host`, `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`.
- `ALL /*`: Reverse proxy tới Frontend service (`http://frontend:3000`), kèm theo cấu hình WebSocket (`Upgrade`, `Connection "upgrade"`) hỗ trợ Hot Reload cho Next.js / Vite.

### 3. Danh mục Điểm Cuối API (API Endpoints Matrix)

| Phân hệ | Phương thức | Endpoint | Mô tả chức năng | Yêu cầu Auth |
| :--- | :--- | :--- | :--- | :--- |
| **Hạ tầng** | `GET` | `/ping` hoặc `/api/ping` | Kiểm tra tình trạng hoạt động của Go Engine | Public |
| **Xác thực** | `POST` | `/api/v1/auth/register` | Đăng ký tài khoản (tạo đồng bộ users, credentials, profile) | Public |
| **Xác thực** | `POST` | `/api/v1/auth/login` | Đăng nhập, cấp JWT tokens (Access/Refresh), phòng brute-force | Public |
| **Xác thực** | `POST` | `/api/v1/auth/refresh` | Cấp mới Access Token bằng Refresh Token | Public |
| **Xác thực** | `GET` | `/api/v1/auth/me` | Lấy thông tin tài khoản đang đăng nhập | Bearer JWT |
| **Xác thực** | `POST` | `/api/v1/auth/logout-all` | Thu hồi toàn bộ phiên đăng nhập (tăng `token_version`) | Bearer JWT |
| **Người dùng** | `GET` | `/api/v1/users` | Danh sách người dùng (Cursor Paging `UUIDv7`, Search, Status) | Public |
| **Người dùng** | `POST` | `/api/v1/users` | Tạo mới người dùng kèm hồ sơ cá nhân qua API | Public |
| **Người dùng** | `GET` | `/api/v1/users/:id` | Xem chi tiết người dùng theo ID (`UUIDv7`) hoặc `public_id` | Public |
| **Người dùng** | `PATCH` / `PUT` | `/api/v1/users/:id` | Cập nhật thông tin tài khoản và hồ sơ/preferences | Public |
| **Người dùng** | `DELETE` | `/api/v1/users/:id` | Xóa người dùng (mặc định soft-delete `status = 'deleted'`) | Public |

---

## ️ Cơ sở Dữ liệu & Migrations (golang-migrate)

Hệ thống sử dụng **PostgreSQL 18** và công cụ **golang-migrate v4** để phiên bản hóa toàn bộ cấu trúc cơ sở dữ liệu.

### 1. Sơ đồ các Bảng Dữ liệu Hiện tại

```mermaid
erDiagram
    users ||--|| user_credentials : "has"
    users ||--|| user_profiles : "has"

    users {
        uuid id PK "UUIDv7"
        varchar public_id UK "Public ID người dùng (hiển thị công khai)"
        varchar email UK "Email người dùng (dùng đăng nhập)"
        varchar full_name "Họ và tên"
        text avatar_url "Ảnh đại diện"
        varchar status "active, inactive, deleted"
        timestamptz created_at
        timestamptz updated_at
    }

    user_credentials {
        uuid user_id PK,FK "References users(id) ON DELETE CASCADE"
        varchar password_hash "Mật khẩu mã hóa (NOT NULL)"
        int token_version "Thu hồi session / refresh token"
        timestamptz last_login_at
        int failed_login_attempts "Chống brute-force"
        timestamptz locked_until "Thời điểm mở khóa"
        timestamptz created_at
        timestamptz updated_at
    }

    user_profiles {
        uuid user_id PK,FK "References users(id) ON DELETE CASCADE"
        date date_of_birth "Ngày sinh"
        varchar gender "Giới tính"
        varchar phone_number "Số điện thoại"
        text address "Địa chỉ"
        jsonb preferences "Tùy chọn JSONB (GIN Indexed)"
        timestamptz created_at
        timestamptz updated_at
    }
```

- **`users`**: Chứa thông tin cốt lõi của tài khoản, sử dụng hàm `uuidv7()` làm khóa chính, trường `public_id` định danh công khai hiển thị cho người dùng, và đánh index cho cột `email` phục vụ đăng nhập.
- **`user_credentials`**: Chứa thông tin bảo mật, xác thực tài khoản; lưu trữ mật khẩu đã được mã hóa (`password_hash NOT NULL`) phục vụ đăng nhập trực tiếp bằng email (từ bảng `users`) và mật khẩu. Quản lý `token_version` để thu hồi phiên/refresh token và các trường chống brute-force đăng nhập (`failed_login_attempts`, `locked_until`).
- **`user_profiles`**: Lưu thông tin mở rộng của người dùng; sử dụng cột `preferences JSONB` được đánh chỉ mục **GIN** (`idx_user_profiles_preferences`) cho truy vấn hiệu năng cao.

### 2. Chuỗi kết nối Database Migration (`DB_URL`)

Khi chạy migration từ bên trong container `terra_server_dev`, chuỗi kết nối mặc định có định dạng:

```bash
postgres://terra_user:terra_secret@postgres:5432/terra_db?sslmode=disable
```

### 3. Thao tác Migrations qua Makefile

```bash
# 1. Tạo migration file mới (cặp file .up.sql và .down.sql)
make migration name=add_orders_table

# 2. Áp dụng tất cả các migration mới lên CSDL
make migrate-up DB_URL="postgres://terra_user:terra_secret@postgres:5432/terra_db?sslmode=disable"

# 3. Rollback (hoàn tác) 1 bước migration gần nhất
make migrate-down DB_URL="postgres://terra_user:terra_secret@postgres:5432/terra_db?sslmode=disable"

# 4. Xóa sạch dữ liệu (drop) và chạy lại toàn bộ migration từ đầu (CẨN THẬN)
make migrate-reset DB_URL="postgres://terra_user:terra_secret@postgres:5432/terra_db?sslmode=disable"
```

> **Mẹo tiện lợi**: Bạn có thể export biến `DB_URL` ra môi trường terminal hoặc cấu hình trực tiếp vào shell để không cần gõ lại tham số mỗi lần:
>
> ```bash
> export DB_URL="postgres://terra_user:terra_secret@postgres:5432/terra_db?sslmode=disable"
> make migrate-up
> ```

---

## 🏗️ Kiến trúc Backend & Tiêu chuẩn Mã nguồn

### 1. Phân tầng Ứng dụng (Layered Clean Architecture)

Hệ thống backend được phân chia rõ ràng trách nhiệm giữa các tầng:

- **`cmd/main.go`**: Điểm khởi động nhẹ nhàng, gọi đến hàm `internal.Run()`.
- **`internal/`**:
  - `config.go`: Tải biến môi trường an toàn bằng `godotenv`, kiểm tra các biến bắt buộc (`PORT`, `DB_HOST`, `DB_NAME`).
  - `postgresql.go`: Quản lý kết nối GORM và thực hiện `Ping()` kiểm tra kết nối với PostgreSQL Pool.
  - `logger.go`: Cấu hình hệ thống ghi log tập trung.
  - `router.go`: Cấu hình router Gin và gắn các handler/middleware.
  - `run.go`: Điều phối toàn bộ vòng đời khởi chạy server và hàm dọn dẹp tài nguyên (`defer cleanupLogger()`).
- **`controllers/`**: Tiếp nhận HTTP request, validate dữ liệu đầu vào và chuyển cho Service.
- **`services/`**: Chứa toàn bộ nghiệp vụ logic, tính toán và điều phối luồng xử lý.
- **`repositories/`**: Trực tiếp tương tác với cơ sở dữ liệu qua GORM/SQL.
- **`middleware/`**: Các chức năng cắt ngang (Cross-cutting concerns) như Authentication, Recovery, Rate Limiting, Request ID,...

### 2. Chuẩn hóa Định dạng API Response (`server/common`)

#### Phản hồi Thành công (`SuccessResponse`)

```json
{
  "data": { ... },
  "paging": {
    "total": 100,
    "cursor": "cursor_token_current",
    "next_cursor": "cursor_token_next"
  },
  "extra": null
}
```

#### Phản hồi Lỗi Chuẩn mực (`DefaultError`)

Cung cấp phân cấp lỗi chi tiết, hỗ trợ lấy StackTrace nội bộ và ẩn các thông tin nhạy cảm khỏi người dùng cuối:

```json
{
  "id": "ERR_NOT_FOUND",
  "code": 404,
  "status": "Not Found",
  "request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
  "message": "Requested resource not found",
  "reason": "User with the specified ID does not exist",
  "details": {
    "field": "id"
  }
}
```

Các hằng số lỗi chuẩn tích hợp sẵn:

- `ErrNotFound` (404)
- `ErrUnauthorized` (401)
- `ErrForbidden` (403)
- `ErrInternalServerError` (500)
- `ErrBadRequest` (400)
- `ErrUnsupportedMediaType` (415)
- `ErrConflict` (409)

### 3. Hệ thống Structured Logging & Xoay vòng File Log

- Sử dụng thư viện chuẩn **`log/slog`** kết hợp **JSON Handler** với tùy chọn `AddSource: true`. Toàn bộ log message (`msg`) bắt buộc viết bằng Tiếng Anh.
- Sử dụng **`lumberjack/v3`** để xoay vòng file log tự động tại `server/logs/app.log`:
  - Kích thước tối đa mỗi file: `10MB`.
  - Số lượng bản backup giữ lại: `5 file`.
  - Thời hạn lưu trữ: `28 ngày`.
  - Tự động nén các file log cũ thành định dạng `.gz`.
- **Hành vi theo môi trường**:
  - `Dev / Local`: Sử dụng `io.MultiWriter` xuất đồng thời ra cả Terminal (`os.Stdout`) và file `logs/app.log` ở cấp độ `LevelInfo`.
  - `Staging / Production`: Chỉ ghi vào file với cấp độ `LevelWarn` nhằm tối ưu hiệu năng I/O.
- Đảm bảo **Zero Log Loss**: Sử dụng cơ chế `defer cleanupLogger()` để flush toàn bộ buffer trong RAM xuống đĩa cứng khi server dừng.

### 4. Quy Chuẩn Chú Thích (Code Comments) & Thông Điệp (Messages)

- **Code Comments**: 
  - **Phạm vi bắt buộc**: (1) **Comment hàm**: Chú thích (docstring) cho toàn bộ các hàm, phương thức và constructor, giải thích rõ chức năng, tham số và giá trị trả về; (2) **Comment đoạn code khó hiểu**: Chú thích giải thích các khối logic phức tạp, thuật toán đặc thù, transaction đa bảng, regex, hoặc edge cases. Tránh comment dư thừa ở các dòng code tự giải thích.
  - **Định dạng**: Mọi dòng chú thích khi viết bắt buộc phải có tiền tố `[VN]` ở đầu và viết bằng tiếng Việt (ví dụ: `// [VN] Khởi tạo kết nối cơ sở dữ liệu`).
- **Messages**: Toàn bộ thông điệp trong hệ thống (log `slog`, API error `message`, `reason`, native Go errors `errors.New`, `fmt.Errorf`, exception/panic string, git commit messages) bắt buộc 100% bằng **Tiếng Anh (English)**.
- **Quy tắc phân định**: Tiền tố `[VN]` CHỈ dùng cho code comment, tuyệt đối không đưa `[VN]` vào nội dung các chuỗi message hệ thống hoặc API response.

### 5. Tiêu Chuẩn GORM Model & Database Migration

- Dự án sử dụng `golang-migrate` làm công cụ quản lý schema cơ sở dữ liệu độc lập, không dùng GORM `AutoMigrate`.
- Toàn bộ struct Entity trong `server/models/` chỉ đóng vai trò Data Mapper: struct tag `gorm` chỉ định nghĩa ánh xạ cột `gorm:"column:<ten_cot>"` (kèm `primaryKey` khi khóa chính không mang tên mặc định `id`, và quan hệ `foreignKey`), loại bỏ toàn bộ các tag DDL như `type:`, `not null`, `default:`, `unique:`.

---

## ⚡ Hướng dẫn Bắt đầu Nhanh (Quick Start)

### 1. Chuẩn bị Môi trường

Sao chép file cấu hình môi trường mẫu:

```bash
# Cấu hình cấp root cho Docker Compose
cp .env.example .env

# Cấu hình riêng cho Golang Server
cp server/.env.example server/.env
```

### 2. Khởi chạy Hệ thống Dev / Local

```bash
# Khởi động Nginx, Go Server và PostgreSQL ở chế độ background
make up

# Kiểm tra trạng thái các container
make ps

# Xem log theo thời gian thực (Server live-reload sẽ hiển thị tại đây)
make logs
```

### 3. Chạy Database Migrations

Áp dụng các bảng dữ liệu ban đầu vào cơ sở dữ liệu:

```bash
make migrate-up DB_URL="postgres://terra_user:terra_secret@postgres:5432/terra_db?sslmode=disable"
```

### 4. Kiểm tra Hoạt động (Health Check)

- **Nginx Gateway**:
  ```bash
  curl http://localhost:8088/healthz
  # Kết quả: healthy
  ```
- **Go Server API qua Nginx**:
  ```bash
  curl http://localhost:8088/api/ping
  # Kết quả: {"message":"pong"}
  ```
- **Go Server Debug Port trực tiếp**:
  ```bash
  curl http://localhost:8090/ping
  # Kết quả: {"message":"pong"}
  ```

---

## 🛠️ Danh mục Lệnh Makefile Toàn diện

Sử dụng lệnh sau để xem danh sách trợ giúp trực tiếp bất kỳ lúc nào:

```bash
make help
```

### 1. Môi trường Dev / Local (Mặc định)

| Lệnh                            | Ý nghĩa                                                             |
| :------------------------------ | :------------------------------------------------------------------ |
| `make up` / `make dev-up`       | Khởi chạy toàn bộ hệ thống dev (Nginx, Server, PostgreSQL)          |
| `make down` / `make dev-down`   | Dừng và tắt các container môi trường dev                            |
| `make build` / `make dev-build` | Build lại toàn bộ container dev khi có cập nhật Dockerfile / go.mod |
| `make logs` / `make dev-logs`   | Theo dõi log theo thời gian thực của tất cả service dev             |
| `make ps` / `make dev-ps`       | Hiển thị danh sách và trạng thái hoạt động của các container dev    |
| `make dev-restart`              | Khởi động lại các container trong môi trường dev                    |

### 2. Môi trường Staging

| Lệnh                 | Ý nghĩa                                   |
| :------------------- | :---------------------------------------- |
| `make staging-up`    | Khởi chạy hệ thống môi trường Staging     |
| `make staging-down`  | Dừng hệ thống môi trường Staging          |
| `make staging-build` | Build lại Docker image môi trường Staging |
| `make staging-logs`  | Theo dõi log môi trường Staging           |

### 3. Môi trường Production

| Lệnh              | Ý nghĩa                                        |
| :---------------- | :--------------------------------------------- |
| `make prod-up`    | Khởi chạy hệ thống môi trường Production       |
| `make prod-down`  | Dừng hệ thống môi trường Production            |
| `make prod-build` | Build lại Docker image hardened cho Production |
| `make prod-logs`  | Theo dõi log môi trường Production             |

### 4. Quản lý Cơ sở Dữ liệu & Migrations

| Lệnh                              | Ý nghĩa                                                              |
| :-------------------------------- | :------------------------------------------------------------------- |
| `make migration name=<tên>`       | Tạo file migration mới theo thứ tự tuần tự trong `server/migrations` |
| `make migrate-up DB_URL="..."`    | Chạy toàn bộ migration mới chưa áp dụng lên PostgreSQL               |
| `make migrate-down DB_URL="..."`  | Rollback (hoàn tác) 1 version migration gần nhất                     |
| `make migrate-reset DB_URL="..."` | Xóa sạch bảng (drop) và migrate lại toàn bộ từ đầu                   |

### 5. Tiện ích Gỡ lỗi & Quản trị Hệ thống

| Lệnh                            | Ý nghĩa                                                                     |
| :------------------------------ | :-------------------------------------------------------------------------- |
| `make server-sh`                | Mở terminal sh trực tiếp bên trong container Go Server (`terra_server_dev`) |
| `make nginx-sh`                 | Mở terminal sh trực tiếp bên trong container Nginx (`terra_nginx_dev`)      |
| `make nginx-reload`             | Nạp lại cấu hình Nginx tức thời mà không làm gián đoạn container            |
| `make psql-cli` / `make db-cli` | Mở giao diện dòng lệnh PostgreSQL CLI (`psql`) trong container database     |
| `make clean`                    | Dừng và xóa sạch toàn bộ container, mạng nội bộ và volumes của dự án        |

---

## 🤖 Quy Trình Phát Triển & Chuẩn Mực Agent (AGENTS.md)

Dự án áp dụng quy chuẩn phát triển tự động hóa nghiêm ngặt đối với AI Agent và các kỹ sư phát triển thông qua tệp quy tắc [`AGENTS.md`](file:///home/wsl-admin/projects/terra-clone/AGENTS.md). 

Mọi tính năng hoặc thay đổi hệ thống đều bắt buộc trải qua quy trình **4 bước tuần tự**:
1. **Lập Kế Hoạch (Planning)**: Bắt buộc khởi tạo thư mục kế hoạch theo định dạng `plans/YYYYMMDD-HHMMSS-<tên-plan>/` chứa tệp [`plan.md`](file:///home/wsl-admin/projects/terra-clone/plans) (nếu một mục tiêu) hoặc các tệp plan con kèm plan tổng hợp `plan.md` (nếu nhiều plan con) trước khi triển khai code.
2. **Nắm Bắt Kiến Trúc & Convention**: Đọc và đối chiếu toàn bộ tài liệu quy chuẩn trong [`docs/`](file:///home/wsl-admin/projects/terra-clone/docs) ([`design-guidelines.md`](file:///home/wsl-admin/projects/terra-clone/docs/design-guidelines.md), [`system-architecture.md`](file:///home/wsl-admin/projects/terra-clone/docs/system-architecture.md), [`codebase-summary.md`](file:///home/wsl-admin/projects/terra-clone/docs/codebase-summary.md),...).
3. **Triển Khai Mã Nguồn**: Thực thi code theo đúng plan, tuân thủ Clean Layered Architecture (`Controller -> Service -> Repository`), xử lý lỗi bằng `common.DefaultError`, migration chuẩn `golang-migrate`, và tuân thủ tuyệt đối quy chuẩn **Code Comment có prefix `[VN]`** cùng **Messages viết bằng Tiếng Anh**.
4. **Đồng Bộ Tài Liệu & Cổng Đóng Task (Mandatory Gate)**: Cập nhật tài liệu tương ứng trong `docs/` (`codebase-summary.md`, `system-architecture.md`, `deployment-guide.md`, `project-roadmap.md`) và `README.md`. **Tuyệt đối không được phép kết thúc task nếu chưa hoàn thành bước này.**

