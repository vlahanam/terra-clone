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
├── plans/                            # Thư mục kế hoạch (cấu trúc YYYYMMDD-HHMMSS-<task>/ chứa plan.md)
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
│   ├── .env.example                  # Mẫu cấu hình môi trường backend (kèm JWT_SECRET)
│   ├── .env                          # Cấu hình môi trường backend đang sử dụng
│   ├── Dockerfile                    # Dockerfile Dev (Tích hợp Air & golang-migrate CLI)
│   ├── Dockerfile.staging            # Dockerfile Staging (Multi-stage build)
│   ├── Dockerfile.product            # Dockerfile Production (Minimal hardened Alpine)
│   ├── go.mod                        # Quản lý dependencies Golang
│   ├── go.sum                        # Checksum hash dependencies
│   ├── cmd/
│   │   └── main.go                   # Điểm khởi chạy chương trình (Entry point)
│   ├── common/                       # Khối dùng chung (Responses, Errors, Pagination, JWT, ID)
│   │   ├── error_response.go         # Chuẩn hóa lỗi API, StackTrace & Error carrier
│   │   ├── id.go                     # Hàm sinh public_id ngẫu nhiên an toàn
│   │   ├── jwt.go                    # Phát sinh và xác thực JWT token pair (Access/Refresh)
│   │   ├── jwt_test.go               # Kiểm thử đơn vị cho JWT và ID generator
│   │   ├── pagination.go             # Cấu trúc phân trang Cursor (Paging)
│   │   └── success_response.go       # Chuẩn hóa phản hồi thành công
│   ├── controllers/                  # Tầng tiếp nhận HTTP Request & trả về Response
│   │   ├── auth_controller.go        # Controller xử lý Đăng ký, Đăng nhập, Refresh, Logout-All, Me
│   │   └── user_controller.go        # Controller xử lý CRUD User, phân trang, lọc và tìm kiếm
│   ├── dto/                          # Khối Data Transfer Objects (Request/Response Binding)
│   │   ├── auth_dto.go               # DTO RegisterRequest, LoginRequest, RefreshTokenRequest, AuthResponse
│   │   └── user_dto.go               # DTO CreateUserRequest, UpdateUserRequest, ListUsersQuery, UserResponse
│   ├── internal/                     # Thư viện và logic vận hành nội bộ server
│   │   ├── app.go                    # Định nghĩa AppConfig struct (kèm JWTSecret)
│   │   ├── config.go                 # Trình nạp và xác thực biến môi trường (Fail-Fast)
│   │   ├── logger.go                 # Hệ thống Structured Logging (slog + lumberjack)
│   │   ├── postgresql.go             # Kết nối GORM & kiểm tra Ping connection pool
│   │   ├── router.go                 # Khởi tạo Gin Engine, gắn middleware & API v1 routes
│   │   └── run.go                    # Điều phối vòng đời và dọn dẹp tài nguyên
│   ├── logs/                         # Thư mục chứa log ứng dụng (tự xoay vòng và nén .gz)
│   │   ├── access.log                # Access log các request HTTP
│   │   └── app.log                   # Application structured log (slog JSON)
│   ├── middleware/                   # Tầng middleware xử lý HTTP
│   │   ├── auth.go                   # Middleware RequireAuth kiểm tra Bearer JWT & token_version
│   │   ├── logger.go                 # Custom Gin Access Logger format chi tiết
│   │   └── request_id.go             # Middleware sinh và gán X-Request-ID cho mỗi request
│   ├── migrations/                   # Các script SQL migration có đánh version
│   │   ├── 000001_users_table.up.sql / .down.sql
│   │   ├── 000002_user_credentials_table.up.sql / .down.sql
│   │   └── 000003_user_profiles_table.up.sql / .down.sql
│   ├── models/                       # Entity models tương tác cơ sở dữ liệu qua GORM
│   │   ├── credential.go             # Entity UserCredentials (password_hash, token_version, lock)
│   │   ├── profile.go                # Entity UserProfile (phone, address, JSONMap preferences)
│   │   └── user.go                   # Entity User (id UUIDv7, public_id, email, status)
│   ├── repositories/                 # Tầng tương tác cơ sở dữ liệu (GORM Data Access)
│   │   ├── credential_repository.go  # Quản lý credentials, đếm failed login, tăng token version
│   │   ├── profile_repository.go     # Quản lý profile người dùng và preferences JSONB
│   │   └── user_repository.go        # CRUD User, tìm kiếm email/public_id, phân trang cursor
│   └── services/                     # Tầng xử lý nghiệp vụ kinh doanh (Business Logic)
│       ├── auth_service.go           # Nghiệp vụ Đăng ký (transaction), Đăng nhập, Refresh, Logout-All
│       └── user_service.go           # Nghiệp vụ CRUD người dùng, phân trang cursor, soft-delete
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
  - Định nghĩa endpoint kiểm tra sức khỏe cơ sở `GET /ping` và `GET /api/ping` trả về `{"message": "pong"}`.

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
  - `ErrNotFound` (404), `ErrUnauthorized` (401), `ErrForbidden` (403), `ErrInternalServerError` (500), `ErrBadRequest` (400), `ErrUnsupportedMediaType` (415), `ErrConflict` (409), `ErrRecordNotFound`. Toàn bộ thông điệp lỗi mặc định được chuẩn hóa 100% bằng Tiếng Anh.

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

#### 4. [`jwt.go`](file:///home/wsl-admin/projects/terra-clone/server/common/jwt.go)
- Quản lý vòng đời JWT Token:
  - Hàm `GenerateTokenPair(secret, userID, publicID, email, tokenVersion)`: Sinh đồng thời Access Token (15 phút) và Refresh Token (7 ngày) theo chuẩn HMAC-SHA256.
  - Hàm `ValidateToken(secret, tokenString, expectedType)`: Kiểm tra tính hợp lệ của chữ ký, thời hạn sống và loại token (`access` hoặc `refresh`).
  - Struct `UserClaims`: Đóng gói `user_id`, `public_id`, `email`, `token_version`, `token_type` và các trường tiêu chuẩn của `jwt.RegisteredClaims`.

#### 5. [`id.go`](file:///home/wsl-admin/projects/terra-clone/server/common/id.go)
- Hàm `GeneratePublicID(prefix)`: Tạo chuỗi định danh ngẫu nhiên mã hóa hex kết hợp tiền tố gợi nhớ (ví dụ: `usr_46587d36c1764cea`) an toàn về mặt mật mã học (`crypto/rand`).

---

### 2.4. Tầng Middleware: [`middleware/`](file:///home/wsl-admin/projects/terra-clone/server/middleware)

#### 1. [`logger.go`](file:///home/wsl-admin/projects/terra-clone/server/middleware/logger.go)
- Hàm [`Logger(output io.Writer)`](file:///home/wsl-admin/projects/terra-clone/server/middleware/logger.go#L12):
  - Trích xuất thông tin truy cập của từng request: Thời gian thực thi (`TimeStamp`), Mã trạng thái có màu sắc trực quan (`StatusCodeColor`), Độ trễ xử lý (`Latency`), Địa chỉ IP máy khách (`ClientIP`), HTTP Method (`MethodColor`), Đường dẫn (`Path`), Thiết bị / Trình duyệt truy cập (`UserAgent`).
  - Ghi vào luồng `accessWriter` đã được cấu hình xoay vòng.

#### 2. [`auth.go`](file:///home/wsl-admin/projects/terra-clone/server/middleware/auth.go)
- Hàm `RequireAuth(jwtSecret, credRepo)`:
  - Trích xuất Header `Authorization: Bearer <token>`.
  - Giải mã và xác thực chữ ký Access Token.
  - Đối chiếu `token_version` trong claims với CSDL để lập tức từ chối token bị thu hồi khi đổi mật khẩu/đăng xuất tất cả thiết bị.
  - Thiết lập các thông số `user_id`, `public_id`, `email`, `claims` vào `gin.Context`.
- Hàm `GetCurrentUserID(c)`: Trích xuất `user_id` (UUID) an toàn từ ngữ cảnh request.

#### 3. [`request_id.go`](file:///home/wsl-admin/projects/terra-clone/server/middleware/request_id.go)
- Hàm `RequestID()`:
  - Trích xuất hoặc phát sinh mã `X-Request-ID` độc nhất theo chuẩn `UUIDv7` cho mỗi request HTTP.
  - Gán vào Header phản hồi và context phục vụ truy vết lỗi end-to-end xuyên suốt hệ thống.

---

### 2.5. Tầng Lược Đồ Dữ Liệu: [`migrations/`](file:///home/wsl-admin/projects/terra-clone/server/migrations)

1. **[`000001_users_table`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000001_users_table.up.sql)**:
   - Tạo bảng `users` với khóa chính `id UUID DEFAULT uuidv7()`.
   - Cột `public_id VARCHAR(50) UNIQUE NOT NULL` đóng vai trò định danh công khai hiển thị cho người dùng, được đánh chỉ mục `idx_users_public_id`.
   - Cột `email` định dạng `VARCHAR(255) UNIQUE` được đánh chỉ mục `idx_users_email` (dùng để đăng nhập).
   - Quản lý trạng thái tài khoản: `status VARCHAR(20) DEFAULT 'active'`.
2. **[`000002_user_credentials_table`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000002_user_credentials_table.up.sql)**:
   - Lưu trữ thông tin xác thực mật khẩu và phiên đăng nhập, tách biệt độc lập với bảng thông tin tài khoản.
   - Liên kết 1:1 với `users(id)` qua khóa ngoại với hành vi `ON DELETE CASCADE`.
   - Lưu trữ `password_hash VARCHAR(255) NOT NULL` phục vụ đăng nhập bằng email và mật khẩu (đã tinh giản, loại bỏ các trường OAuth chưa sử dụng ở giai đoạn này).
   - Quản lý phiên (`token_version`), lịch sử đăng nhập (`last_login_at`) và phòng vệ brute-force (`failed_login_attempts`, `locked_until`).
3. **[`000003_user_profiles_table`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000003_user_profiles_table.up.sql)**:
   - Lưu trữ thông tin chi tiết: `date_of_birth`, `gender`, `phone_number`, `address`.
   - Cột `preferences JSONB` lưu cấu hình tùy biến của người dùng, được đánh chỉ mục GIN (`idx_user_profiles_preferences`).

---

### 2.6. Tầng Thực Thể Cơ Sở Dữ Liệu: [`models/`](file:///home/wsl-admin/projects/terra-clone/server/models)
- Toàn bộ các model tuân thủ quy chuẩn tích hợp cùng `golang-migrate`: Struct tag GORM chỉ định nghĩa ánh xạ cột `gorm:"column:<ten_cot>"` (kèm `primaryKey` cho khóa chính và quan hệ `foreignKey`), loại bỏ toàn bộ các tag DDL dư thừa của GORM AutoMigrate (`type:`, `not null`, `default:`, `unique`, `autoCreateTime`, `autoUpdateTime`).
- [`user.go`](file:///home/wsl-admin/projects/terra-clone/server/models/user.go): Khai báo Entity `User` ánh xạ tới bảng `users`, liên kết quan hệ 1:1 với `UserCredentials` và `UserProfile`.
- [`credential.go`](file:///home/wsl-admin/projects/terra-clone/server/models/credential.go): Khai báo Entity `UserCredentials` ánh xạ tới `user_credentials`, bổ sung phương thức kiểm tra trạng thái khóa `IsLocked()`.
- [`profile.go`](file:///home/wsl-admin/projects/terra-clone/server/models/profile.go): Khai báo Entity `UserProfile` và kiểu tùy chỉnh `JSONMap` hiện thực `sql.Scanner` và `driver.Valuer` phục vụ đọc/ghi cột `preferences JSONB`.

---

### 2.7. Tầng Truy Cập Dữ Liệu: [`repositories/`](file:///home/wsl-admin/projects/terra-clone/server/repositories)
- [`user_repository.go`](file:///home/wsl-admin/projects/terra-clone/server/repositories/user_repository.go): Đóng gói toàn bộ câu lệnh truy vấn bảng `users`: `Create`, `FindByID`, `FindByPublicID`, `FindByEmail`, `FindByIDOrPublicID`, `List` (hỗ trợ phân trang con trỏ `UUIDv7`), `Update`, `Delete` (soft-delete hoặc hard-delete).
- [`credential_repository.go`](file:///home/wsl-admin/projects/terra-clone/server/repositories/credential_repository.go): Quản lý bảng `user_credentials`: `Create`, `FindByUserID`, `Update`, `RecordFailedLogin` (tự động khóa 15 phút khi sai liên tiếp 5 lần), `ResetFailedLogin`, `IncrementTokenVersion`.
- [`profile_repository.go`](file:///home/wsl-admin/projects/terra-clone/server/repositories/profile_repository.go): Quản lý bảng `user_profiles`: `Create`, `FindByUserID`, `Update`, `Upsert` (sử dụng mệnh đề `ON CONFLICT DO UPDATE`).

---

### 2.8. Tầng Xử Lý Nghiệp Vụ: [`services/`](file:///home/wsl-admin/projects/terra-clone/server/services)
- [`auth_service.go`](file:///home/wsl-admin/projects/terra-clone/server/services/auth_service.go):
  - `Register`: Đăng ký tài khoản trong transaction 3 bảng (`users`, `user_credentials`, `user_profiles`), mã hóa `bcrypt`, phát sinh JWT Token Pair.
  - `Login`: Xác thực tài khoản, kiểm tra chống brute-force `failed_login_attempts` và `locked_until`, reset trạng thái khi đăng nhập thành công, phát sinh JWT token.
  - `RefreshToken`: Xác thực refresh token, đối chiếu `token_version` và cấp mới cặp token.
  - `LogoutAll`: Tăng `token_version` để thu hồi tức thời toàn bộ phiên đăng nhập.
  - `GetMe`: Lấy thông tin chi tiết người dùng đang đăng nhập.
- [`user_service.go`](file:///home/wsl-admin/projects/terra-clone/server/services/user_service.go):
  - `CreateUser`: Tạo mới người dùng kèm hồ sơ cá nhân qua API quản trị.
  - `GetUserByID`: Tra cứu người dùng theo ID nội bộ hoặc `public_id`.
  - `ListUsers`: Liệt kê danh sách người dùng với con trỏ `UUIDv7`, tìm kiếm theo email/họ tên và lọc trạng thái.
  - `UpdateUser`: Cập nhật thông tin tài khoản và upsert thông tin hồ sơ/preferences.
  - `DeleteUser`: Xóa người dùng (mặc định soft-delete `status = 'deleted'`).

---

### 2.9. Tầng Giao Tiếp Vận Chuyển: [`controllers/`](file:///home/wsl-admin/projects/terra-clone/server/controllers)
- [`auth_controller.go`](file:///home/wsl-admin/projects/terra-clone/server/controllers/auth_controller.go): Điều phối các endpoint `/api/v1/auth/*`: `Register`, `Login`, `RefreshToken`, `LogoutAll`, `GetMe`.
- [`user_controller.go`](file:///home/wsl-admin/projects/terra-clone/server/controllers/user_controller.go): Điều phối các endpoint `/api/v1/users/*`: `Create`, `GetByID`, `List`, `Update`, `Delete`.

---

### 2.10. Tầng Data Transfer Objects: [`dto/`](file:///home/wsl-admin/projects/terra-clone/server/dto)
- [`auth_dto.go`](file:///home/wsl-admin/projects/terra-clone/server/dto/auth_dto.go): Định nghĩa cấu trúc binding JSON cho luồng xác thực (`RegisterRequest`, `LoginRequest`, `RefreshTokenRequest`, `AuthResponse`).
- [`user_dto.go`](file:///home/wsl-admin/projects/terra-clone/server/dto/user_dto.go): Định nghĩa cấu trúc binding và serialization cho quản trị người dùng (`CreateUserRequest`, `UpdateUserRequest`, `ListUsersQuery`, `UserProfileResponse`, `UserResponse`).

---

### 2.11. Quy Chuẩn Chú Thích & Thông Điệp (Language & Comment Conventions)
- **Code Comments**:
  - **Phạm vi bắt buộc**: (1) Comment hàm (docstring mô tả trách nhiệm, tham số, kết quả); (2) Comment các đoạn code khó hiểu (thuật toán phức tạp, transaction, edge cases, workaround). Tránh comment rườm rà ở những dòng code hiển nhiên.
  - **Định dạng**: Mọi chú thích mã nguồn bắt buộc có tiền tố `[VN]` ở đầu và viết bằng tiếng Việt.
- **Messages**: Tất cả các loại thông điệp (messages) trong hệ thống bắt buộc viết bằng Tiếng Anh (English):
  - Log messages trong `slog` (ví dụ: `slog.Info("server started successfully", ...)`).
  - Error messages (`ErrorField`, `ReasonField` trong `common.DefaultError`, chuỗi trong `errors.New`, `fmt.Errorf`). Toàn bộ chuỗi lý do lỗi (`WithReason`, `WithReasonf`) tại các tầng `controllers/`, `middleware/`, `services/` đã được chuẩn hóa 100% bằng Tiếng Anh.
  - Phản hồi HTTP API hoặc status text.
  - Commit messages trong Git.
- **Phân định rõ ràng**: Tiền tố `[VN]` chỉ áp dụng cho chú thích mã nguồn, tuyệt đối không đặt tiền tố `[VN]` vào nội dung chuỗi message của ứng dụng.
