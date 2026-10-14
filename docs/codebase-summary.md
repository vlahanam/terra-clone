# 💻 Tóm Tắt Mã Nguồn Dự Án (Codebase Summary)
## Dự án: Terra Clone

---

## 1. 🗂️ Cây Thư Mục Tổng Thể (Directory Tree)

```
terra-clone/
├── .dockerignore                     # Loại trừ file không cần thiết khi build Docker root
├── .env.example                      # Mẫu biến môi trường cấp root (Compose)
├── .env                              # Biến môi trường thực tế đang áp dụng
├── AGENTS.md                         # Quy chuẩn & quy trình phát triển bắt buộc cho AI Agent
├── docker-compose.yml                # Cấu hình container cho Dev / Local (Air live-reload)
├── docker-compose.staging.yml        # Cấu hình container cho Staging
├── docker-compose.product.yml        # Cấu hình container cho Production
├── Makefile                          # Danh mục lệnh CLI tự động hóa
├── plans/                            # Kế hoạch chi tiết từng chức năng trước khi code
├── docs/                             # Tài liệu kiến trúc và hướng dẫn kỹ thuật
│   ├── codebase-summary.md           # [Tài liệu hiện tại] Bóc tách toàn bộ mã nguồn
│   ├── deployment-guide.md           # Cẩm nang triển khai đa môi trường
│   ├── design-guidelines.md          # Bộ quy chuẩn thiết kế và tiêu chuẩn code
│   ├── project-overview-pdr.md       # Yêu cầu thiết kế sản phẩm (PDR)
│   ├── project-roadmap.md            # Kế hoạch phát triển các giai đoạn
│   └── system-architecture.md        # Kiến trúc kỹ thuật và luồng dữ liệu
├── nginx/                            # Cấu hình Gateway và Reverse Proxy
│   ├── Dockerfile                    # Image Nginx môi trường Dev
│   ├── Dockerfile.staging            # Image Nginx môi trường Staging
│   ├── Dockerfile.product            # Image Nginx môi trường Production
│   ├── nginx.conf                    # Cấu hình cốt lõi Nginx (Gzip, buffers, log format)
│   └── conf.d/
│       ├── default.conf              # Virtual host cho Dev / Local
│       ├── staging.conf              # Virtual host cho Staging
│       └── product.conf              # Virtual host cho Production (Security headers)
├── server/                           # Backend Application Server (Golang 1.26+)
│   ├── .air.toml                     # Cấu hình live-reloading với Air
│   ├── .dockerignore                 # Loại trừ file khi build image Go Server
│   ├── .env.example                  # Mẫu cấu hình môi trường backend
│   ├── .env                          # Cấu hình môi trường backend đang sử dụng
│   ├── Dockerfile                    # Dockerfile Dev (Tích hợp Air & golang-migrate CLI)
│   ├── Dockerfile.staging            # Dockerfile Staging (Multi-stage build)
│   ├── Dockerfile.product            # Dockerfile Production (Minimal hardened Alpine)
│   ├── go.mod                        # Quản lý dependencies Golang
│   ├── go.sum                        # Checksum hash dependencies
│   ├── cmd/
│   │   └── main.go                   # Điểm khởi chạy chương trình (Entry point)
│   ├── common/                       # Khối dùng chung (Responses, Errors, Pagination)
│   │   ├── error_response.go         # Chuẩn hóa lỗi API, StackTrace & Error carrier
│   │   ├── pagination.go             # Cấu trúc phân trang Cursor (Paging)
│   │   └── success_response.go       # Chuẩn hóa phản hồi thành công
│   ├── controllers/                  # Tầng tiếp nhận HTTP Request & trả về Response
│   ├── internal/                     # Thư viện và logic vận hành nội bộ server
│   │   ├── app.go                    # Định nghĩa AppConfig struct
│   │   ├── config.go                 # Trình nạp và xác thực biến môi trường
│   │   ├── logger.go                 # Hệ thống Structured Logging (slog + lumberjack)
│   │   ├── postgresql.go             # Kết nối GORM & kiểm tra Ping connection pool
│   │   ├── router.go                 # Khởi tạo Gin Engine, gắn middleware & routes
│   │   └── run.go                    # Điều phối vòng đời và dọn dẹp tài nguyên
│   ├── logs/                         # Thư mục chứa log ứng dụng (tự xoay vòng và nén .gz)
│   │   ├── access.log                # Access log các request HTTP
│   │   └── app.log                   # Application structured log (slog JSON)
│   ├── middleware/                   # Tầng middleware xử lý HTTP
│   │   └── logger.go                 # Custom Gin Access Logger format chi tiết
│   ├── migrations/                   # Các script SQL migration có đánh version
│   │   ├── 000001_users_table.up.sql / .down.sql
│   │   ├── 000002_user_credentials_table.up.sql / .down.sql
│   │   └── 000003_user_profiles_table.up.sql / .down.sql
│   ├── repositories/                 # Tầng tương tác cơ sở dữ liệu (GORM)
│   └── services/                     # Tầng xử lý nghiệp vụ kinh doanh (Business Logic)
└── frontend/                         # Thư mục mã nguồn Frontend Web (Next.js)
    └── .dockerignore                 # File ignore cho Docker frontend
```

---

## 2. 🔍 Bóc Tách Chi Tiết Từng Module Backend (`server/`)

### 2.1. Tầng Entrypoint: [`cmd/main.go`](file:///home/wsl-admin/projects/terra-clone/server/cmd/main.go)
- **Nhiệm vụ**: Đóng vai trò là entrypoint tối giản nhất có thể.
- **Hiện thực**: Gọi duy nhất hàm [`internal.Run()`](file:///home/wsl-admin/projects/terra-clone/server/internal/run.go#L8). Toàn bộ logic khởi tạo được chuyển giao cho gói `internal` quản lý nhằm đảm bảo tính đóng gói và dễ viết test tích hợp.

### 2.2. Gói Vận Hành Nội Bộ: [`internal/`](file:///home/wsl-admin/projects/terra-clone/server/internal)

#### 1. [`app.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/app.go)
Định nghĩa cấu trúc ứng dụng:
- Struct [`AppConfig`](file:///home/wsl-admin/projects/terra-clone/server/internal/app.go#L3-L6): Chứa 2 trường cơ sở `AppEnv` (`local`, `staging`, `production`) và `Port` (cổng lắng nghe của server).

#### 2. [`config.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/config.go)
- Hàm [`NewConfig()`](file:///home/wsl-admin/projects/terra-clone/server/internal/config.go#L15):
  - Sử dụng thư viện `godotenv` nạp file `.env`. Nếu không có file `.env`, hệ thống in cảnh báo thân thiện và tự động đọc từ biến môi trường hệ thống (System Environment).
  - Khởi tạo `AppConfig` và `DBConfig`.
  - Kiểm tra điều kiện bắt buộc: nếu thiếu `Port`, `DB_HOST`, hoặc `DB_NAME` thì trả về lỗi ngay lập tức để dừng khởi động server (Fail-fast principle).

#### 3. [`postgresql.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/postgresql.go)
- Struct [`DBConfig`](file:///home/wsl-admin/projects/terra-clone/server/internal/postgresql.go#L10-L17): Lưu thông tin kết nối CSDL (`Host`, `Port`, `Name`, `Username`, `Password`, `Sslmode`).
- Phương thức `DSN()`: Tạo chuỗi Data Source Name chuẩn PostgreSQL với timezone UTC.
- Phương thức `ConnectDb()`:
  - Mở kết nối GORM thông qua driver `postgres`.
  - Lấy đối tượng `sql.DB` bên dưới và thực hiện `Ping()` kiểm tra kết nối mạng thực tế tới PostgreSQL Server.

#### 4. [`logger.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/logger.go)
- Khởi tạo hệ thống ghi log tập trung bằng `log/slog` kết hợp `lumberjack/v3`:
  - `fileRoller`: Ghi ra file `logs/app.log`.
  - `accessRoller`: Ghi ra file `logs/access.log`.
  - Các hằng số vận hành: `maxSizeBytes = 10MB`, `maxBackups = 5`, `maxAgeDuration = 28 ngày`, `compressOld = true` (nén `.gz`).
- **Phân tách hành vi theo môi trường**:
  - `local / dev`: Ghi đồng thời ra terminal (`os.Stdout`) và file roller (`io.MultiWriter`), cấp độ `LevelInfo`.
  - `staging / production`: Chỉ ghi ra file roller để tránh nghẽn I/O terminal, cấp độ `LevelWarn`.
- Hàm `cleanup()`: Đóng 2 roller, đẩy sạch bộ đệm ghi đĩa trước khi shutdown server.

#### 5. [`router.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/router.go)
- Hàm [`SetupRouter()`](file:///home/wsl-admin/projects/terra-clone/server/internal/router.go#L11):
  - Tự động chuyển Gin sang chế độ `ReleaseMode` khi ở môi trường Staging/Production.
  - Khởi tạo `gin.New()` thuần túy.
  - Gắn middleware ghi access log [`middleware.Logger(accessWriter)`](file:///home/wsl-admin/projects/terra-clone/server/middleware/logger.go#L12) và [`gin.Recovery()`](file:///home/wsl-admin/projects/terra-clone/server/internal/router.go#L17) ngăn sập tiến trình khi có panic.
  - Định nghĩa endpoint kiểm tra sức khỏe cơ sở `GET /ping` trả về `{"message": "pong"}`.

#### 6. [`run.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/run.go)
- Hàm [`Run()`](file:///home/wsl-admin/projects/terra-clone/server/internal/run.go#L8):
  - Nạp cấu hình qua `NewConfig()`.
  - Khởi tạo Logger và thiết lập `defer cleanupLogger()`.
  - Khởi tạo kết nối CSDL qua `ConnectDb()`.
  - Thiết lập Router và lắng nghe trên cổng cấu hình (mặc định `:8080`).

---

### 2.3. Tầng Dùng Chung: [`common/`](file:///home/wsl-admin/projects/terra-clone/server/common)

#### 1. [`error_response.go`](file:///home/wsl-admin/projects/terra-clone/server/common/error_response.go)
Cung cấp bộ công cụ quản lý lỗi cấp doanh nghiệp:
- Struct [`DefaultError`](file:///home/wsl-admin/projects/terra-clone/server/common/error_response.go#L12-L53):
  - `IDField` (`id`): Mã lỗi logic (ví dụ: `ERR_NOT_FOUND`, `ERR_UNAUTHORIZED`).
  - `CodeField` (`code`): HTTP Status Code tương ứng (ví dụ: `404`, `401`, `500`).
  - `StatusField` (`status`): Mô tả text ngắn của HTTP Status.
  - `RIDField` (`request`): Request ID truy vết sự cố.
  - `ReasonField` (`reason`): Lý do nghiệp vụ thân thiện với người dùng cuối.
  - `DebugField` (`debug`): Chi tiết lỗi kỹ thuật chỉ dành cho đội ngũ phát triển (không serialize cho client).
  - `ErrorField` (`message`): Thông điệp tóm lược lỗi chính.
  - `DetailsField` (`details`): Map chứa thông tin chi tiết lỗi (ví dụ: tên trường dữ liệu không hợp lệ).
  - `err error`: Đối tượng lỗi gốc được bọc lại (wrapped error) để lấy stacktrace qua `github.com/pkg/errors`.
- Hằng số lỗi định sẵn:
  - `ErrNotFound` (404), `ErrUnauthorized` (401), `ErrForbidden` (403), `ErrInternalServerError` (500), `ErrBadRequest` (400), `ErrUnsupportedMediaType` (415), `ErrConflict` (409).

#### 2. [`success_response.go`](file:///home/wsl-admin/projects/terra-clone/server/common/success_response.go)
Chuẩn hóa dữ liệu trả về cho tất cả các API thành công:
```json
{
  "data": { ... },
  "paging": { ... },
  "extra": { ... }
}
```
- Hàm [`SuccessResponse(data, paging, extra)`](file:///home/wsl-admin/projects/terra-clone/server/common/success_response.go#L9).
- Hàm [`ResponseData(data)`](file:///home/wsl-admin/projects/terra-clone/server/common/success_response.go#L13) dùng khi không cần phân trang.

#### 3. [`pagination.go`](file:///home/wsl-admin/projects/terra-clone/server/common/pagination.go)
- Struct [`Paging`](file:///home/wsl-admin/projects/terra-clone/server/common/pagination.go#L3-L7):
  - Áp dụng kỹ thuật **Cursor-based Pagination** (thay thế cho Offset-based truyền thống), tối ưu hóa hiệu năng truy vấn trên bảng dữ liệu lớn hàng chục triệu bản ghi kết hợp với khóa chính `UUIDv7`.

---

### 2.4. Tầng Middleware: [`middleware/logger.go`](file:///home/wsl-admin/projects/terra-clone/server/middleware/logger.go)
- Hàm [`Logger(output io.Writer)`](file:///home/wsl-admin/projects/terra-clone/server/middleware/logger.go#L12):
  - Trích xuất thông tin truy cập của từng request: Thời gian thực thi (`TimeStamp`), Mã trạng thái có màu sắc trực quan (`StatusCodeColor`), Độ trễ xử lý (`Latency`), Địa chỉ IP máy khách (`ClientIP`), HTTP Method (`MethodColor`), Đường dẫn (`Path`), Thiết bị / Trình duyệt truy cập (`UserAgent`).
  - Ghi vào luồng `accessWriter` đã được cấu hình xoay vòng.

---

### 2.5. Tầng Lược Đồ Dữ Liệu: [`migrations/`](file:///home/wsl-admin/projects/terra-clone/server/migrations)

1. **[`000001_users_table`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000001_users_table.up.sql)**:
   - Tạo bảng `users` với khóa chính `id UUID DEFAULT uuidv7()`.
   - Cột `email` định dạng `VARCHAR(255) UNIQUE` được đánh chỉ mục `idx_users_email`.
   - Quản lý trạng thái tài khoản: `status VARCHAR(20) DEFAULT 'active'`.
2. **[`000002_user_credentials_table`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000002_user_credentials_table.up.sql)**:
   - Lưu trữ thông tin đăng nhập, tách biệt độc lập với bảng hồ sơ cá nhân.
   - Liên kết 1:1 với `users(id)` qua khóa ngoại với hành vi `ON DELETE CASCADE`.
   - Hỗ trợ cả đăng nhập `local` và `OAuth` (Google, v.v.) qua ràng buộc:
     `CONSTRAINT chk_local_password CHECK (auth_provider != 'local' OR (username IS NOT NULL AND password_hash IS NOT NULL))`
   - Đánh chỉ mục tìm kiếm `idx_user_credentials_provider` và `idx_user_credentials_username`.
3. **[`000003_user_profiles_table`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000003_user_profiles_table.up.sql)**:
   - Lưu trữ thông tin chi tiết: `date_of_birth`, `gender`, `phone_number`, `address`.
   - Cột `preferences JSONB` lưu cấu hình tùy biến của người dùng, được đánh chỉ mục GIN (`idx_user_profiles_preferences`).
