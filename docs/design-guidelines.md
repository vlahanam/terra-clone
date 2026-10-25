# 📐 Quy Chuẩn Thiết Kế & Tiêu Chuẩn Mã Nguồn (Design Guidelines)
## Dự án: Terra Clone

---

## 1. 🎯 Nguyên Tắc Kiến Trúc Cốt Lõi (Core Principles)

Dự án **Terra Clone** tuân thủ các nguyên lý thiết kế công nghệ phần mềm hiện đại nhằm đảm bảo khả năng mở rộng, tính bảo trì và độ tin cậy:

### 1.1. Kiến Trúc Phân Tầng Độc Lập (Layered Clean Architecture)
Mã nguồn backend tuân thủ quy tắc một chiều nghiêm ngặt:
$$\text{Transport (Controller)} \longrightarrow \text{Domain (Service)} \longrightarrow \text{Data (Repository)}$$

- **Tầng Controllers (`controllers/`)**:
  - Chỉ đóng vai trò tiếp nhận HTTP Request từ `gin.Context`.
  - Phân tích và validate dữ liệu đầu vào (DTO Binding & Validation).
  - Gọi xuống tầng Service tương ứng.
  - Trả về HTTP Response sử dụng các hàm chuẩn [`common.SuccessResponse`](file:///home/wsl-admin/projects/terra-clone/server/common/success_response.go) hoặc [`common.DefaultError`](file:///home/wsl-admin/projects/terra-clone/server/common/error_response.go).
  - **Quy tắc cấm**: Tuyệt đối không viết logic nghiệp vụ (business logic) hoặc gọi trực tiếp tầng Repository/CSDL tại Controller.

- **Tầng Services (`services/`)**:
  - Chứa toàn bộ logic nghiệp vụ, tính toán, phân quyền và điều phối quy trình xử lý.
  - Phối hợp giữa nhiều Repository khác nhau hoặc gọi các dịch vụ bên thứ ba (Third-party APIs).
  - **Quy tắc cấm**: Tầng Service không được phụ thuộc vào `gin.Context` hoặc bất kỳ đối tượng nào thuộc tầng HTTP Transport. Service chỉ nhận và trả về struct Go thuần túy (`context.Context`, DTOs, errors).

- **Tầng Repositories (`repositories/`)**:
  - Đóng gói toàn bộ câu lệnh truy vấn cơ sở dữ liệu qua GORM hoặc Raw SQL.
  - Chịu trách nhiệm map kết quả từ CSDL vào Entity struct.
  - **Quy tắc cấm**: Không đặt các điều kiện rẽ nhánh nghiệp vụ kinh doanh tại tầng Repository.

### 1.2. Nguyên Tắc Thất Bại Sớm (Fail-Fast Principle)
Hệ thống phải xác thực cấu hình và kiểm tra kết nối với các hạ tầng quan trọng ngay trong giai đoạn khởi động:
- Kiểm tra các biến môi trường bắt buộc trong [`internal/config.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/config.go#L34-L36). Nếu thiếu thông số thiết yếu (`PORT`, `DB_HOST`, `DB_NAME`), ứng dụng lập tức trả về lỗi và dừng tiến trình.
- Thực hiện `Ping()` kiểm tra kết nối CSDL trong [`internal/postgresql.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/postgresql.go#L37-L39). Nếu CSDL không phản hồi, tiến trình không khởi chạy server.

---

## 2. 📝 Tiêu Chuẩn Viết Mã Nguồn Golang (Go Coding Standards)

### 2.1. Quy Ước Đặt Tên (Naming Conventions)
- **Tên Gói (Package Names)**: Viết thường hoàn toàn (`lowercase`), ngắn gọn, không dấu gạch dưới hoặc camelCase (ví dụ: `controllers`, `services`, `middleware`, `internal`, `common`).
- **Tên Biến & Hàm**:
  - Public / Exported: Sử dụng PascalCase (ví dụ: `AppConfig`, `NewConfig`, `ConnectDb`, `SetupRouter`).
  - Private / Internal: Sử dụng camelCase (ví dụ: `appConfig`, `dbConfig`, `fileRoller`, `cleanup`).
- **Struct Tags**:
  - Tag JSON: Sử dụng `snake_case` (ví dụ: `json:"full_name"`, `json:"created_at,omitempty"`).
  - Tag Form binding: `form:"cursor"`.
  - Tag GORM: Do hệ thống sử dụng `golang-migrate` để quản lý schema migration CSDL độc lập và không dùng GORM `AutoMigrate`, các model GORM chỉ cần khai báo ánh xạ cột `gorm:"column:ten_cot"` (kèm `primaryKey` khi khóa chính không mang tên mặc định `id`, và quan hệ `foreignKey`), loại bỏ toàn bộ các tag DDL như `type:`, `not null`, `default:`, `unique:`. Ví dụ: `gorm:"column:full_name"`.

### 2.2. Quy Tắc Xử Lý Lỗi (Error Handling Guidelines)
1. **Không nuốt lỗi**: Tuyệt đối không gán `_ = err` khi gọi các thao tác có khả năng sinh lỗi nghiêm trọng.
2. **Bọc lỗi giữ nguyên StackTrace**:
   - Sử dụng [`common.DefaultError`](file:///home/wsl-admin/projects/terra-clone/server/common/error_response.go) kết hợp phương thức `.WithTrace(err)` để đóng gói ngữ cảnh lỗi gốc từ thư viện `github.com/pkg/errors`.
   - Mẫu xử lý chuẩn:
     ```go
     user, err := repo.FindUserByID(ctx, id)
     if err != nil {
         return nil, common.ErrNotFound.
             WithReason("User does not exist in the system").
             WithTrace(err)
     }
     ```
3. **Phân tách thông tin nhạy cảm**:
   - Trường `ReasonField` / `ErrorField`: Chứa thông báo an toàn, thân thiện để gửi tới người dùng (viết bằng Tiếng Anh).
   - Trường `DebugField` và `StackTrace()`: Chỉ ghi vào file log nội bộ hoặc phục vụ debug ở môi trường `local`, không serialize ra payload JSON gửi tới client ở môi trường `production`.

### 2.3. Quy Chuẩn Chú Thích (Code Comments) & Thông Điệp (Messages)

Nhằm đảm bảo sự cân bằng giữa tính dễ hiểu, khả năng bảo trì và tránh gây nhiễu mã nguồn:

#### 1. Phạm vi bắt buộc viết Chú thích (When to Comment):
- **Comment hàm (Functions / Methods / Constructors)**:
  - Bắt buộc viết chú thích (docstring) cho toàn bộ các hàm, phương thức, constructor.
  - Chú thích phải nêu rõ: chức năng của hàm, trách nhiệm xử lý, ý nghĩa các tham số đầu vào và kết quả trả về.
- **Comment các đoạn code khó hiểu (Complex / Non-trivial Logic)**:
  - Bắt buộc giải thích lý do và cách thức xử lý tại các khối code phức tạp: thuật toán đặc thù, database transaction liên quan nhiều bảng, regex pattern, biểu thức điều kiện phức tạp, các xử lý biên (edge cases) hoặc workaround kỹ thuật.
- **Tránh comment dư thừa (Avoid Redundant Comments)**:
  - Tuyệt đối không viết comment vào các dòng code hiển nhiên, tự giải thích (self-explanatory) như khai báo biến đơn giản, `return err`, tăng biến đếm, hoặc các getter/setter tầm thường.

#### 2. Quy chuẩn Định dạng Chú thích (Code Comment Format):
- **Bắt buộc tiền tố `[VN]`**: Mọi dòng chú thích khi viết trong mã nguồn (Golang `//`, `/* */`, SQL `--`, YAML `#`, Dockerfile `#`, Makefile `#`, Shell Script `#`) **BẮT BUỘC** phải có tiền tố `[VN]` ngay sau ký hiệu chú thích.
- Chú thích được viết bằng tiếng Việt rõ nghĩa, mô tả mục đích, nghiệp vụ hoặc lý do kỹ thuật.
- **Ví dụ đúng (Do)**:
  ```go
  // [VN] ConnectDb thực hiện kết nối tới database và kiểm tra ping theo nguyên tắc Fail-Fast
  func (c *DBConfig) ConnectDb() (*gorm.DB, error) {
      db, err := gorm.Open(postgres.Open(c.DSN()), &gorm.Config{})
      if err != nil {
          return nil, fmt.Errorf("failed to open database connection: %w", err)
      }

      sqlDB, err := db.DB()
      if err != nil {
          return nil, fmt.Errorf("failed to get sql.DB instance: %w", err)
      }

      // [VN] Kiểm tra kết nối thực tế tới CSDL để đảm bảo fail-fast ngay khi khởi động
      if err := sqlDB.Ping(); err != nil {
          return nil, fmt.Errorf("failed to ping database: %w", err)
      }

      return db, nil
  }
  ```
- **Ví dụ sai (Don't)**:
  ```go
  // Connect to database (Sai vì thiếu tiền tố [VN])
  // Khởi tạo database (Sai vì thiếu tiền tố [VN])
  // [VN] Trả về db và nil (Sai vì comment dòng code hiển nhiên, dư thừa)
  return db, nil
  ```

#### 3. Quy chuẩn Thông điệp (Messages):
- **Bắt buộc viết bằng Tiếng Anh (English)**: Toàn bộ thông điệp trong hệ thống **BẮT BUỘC 100% bằng Tiếng Anh**. Bao gồm:
  - Thông điệp ghi log qua `slog` (`slog.Info("database connection established")`, `slog.Error("failed to query user", ...)`).
  - Thông báo lỗi API (`ErrorField` trong `common.DefaultError`, ví dụ: `"Requested resource not found"`).
  - Lý do chi tiết của lỗi (`ReasonField` trong `common.DefaultError`, ví dụ: `"User with the given ID does not exist"`).
  - Chuỗi lỗi Go (`errors.New("record not found")`, `fmt.Errorf("failed to parse token: %w", err)`).
  - Thông điệp ngoại lệ (`panic("failed to initialize logger")`).
  - Git commit messages.
- **LƯU Ý ĐẶC BIỆT (Strict Separation)**:
  - Tiền tố `[VN]` **CHỈ dành riêng cho code comment**.
  - **TUYỆT ĐỐI CẤM** đưa tiền tố `[VN]` vào nội dung chuỗi message của mã nguồn (ví dụ: cấm `errors.New("[VN] Record not found")` hoặc `ErrorField: "[VN] Not found"`).

---

## 3. 🌐 Tiêu Chuẩn Thiết Kế RESTful API (API Design Standards)

### 3.1. Quy Ước Định Tuyến (URL Conventions)
- Tiền tố API: `/api/v1/...`
- Sử dụng danh từ số nhiều, chữ thường, phân cách bằng dấu gạch ngang (`kebab-case`):
  - `GET    /api/v1/users` (Lấy danh sách người dùng)
  - `POST   /api/v1/users` (Tạo tài khoản mới)
  - `GET    /api/v1/users/:id` (Xem thông tin chi tiết một người dùng)
  - `PATCH  /api/v1/users/:id` (Cập nhật một phần hồ sơ)
  - `DELETE /api/v1/users/:id` (Vô hiệu hóa hoặc xóa người dùng)
  - `POST   /api/v1/auth/login` (Đăng nhập)
  - `POST   /api/v1/auth/logout-all` (Thu hồi tất cả phiên qua tăng `token_version`)

### 3.2. Chuẩn Hóa Payload Phản Hồi Thành Công
Toàn bộ API thành công phải trả về cấu trúc tuân theo [`common.successResponse`](file:///home/wsl-admin/projects/terra-clone/server/common/success_response.go):
```json
{
  "data": {
    "id": "usr_018f2a1b9f3b7f28",
    "email": "user@example.com",
    "full_name": "Nguyễn Văn A",
    "status": "active"
  },
  "paging": {
    "total": 1250,
    "cursor": "usr_018f2a1b9f3b7f28",
    "next_cursor": "usr_018f2a1c4b127e90"
  },
  "extra": null
}
```

> [!IMPORTANT]
> **Quy Chuẩn Bảo Mật Định Danh (Public Identifier Mapping to `id`)**:
> Trường định danh trong JSON Response được đặt tag `json:"id"` (hoặc `json:"user_id"`, `json:"company_id"` trong bảng liên kết), nhưng **giá trị trả về thực chất luôn là mã định danh công khai (`public_id`)** (ví dụ: `usr_...`, `cmp_...`, `ucmp_...`). Khóa chính UUIDv7 nội bộ tuyệt đối không bao giờ để lộ ra ngoài payload phản hồi (`data`, `paging`, `details`). Phân trang con trỏ `cursor` / `next_cursor` cũng chỉ sử dụng chuỗi `public_id`.

### 3.3. Chuẩn Hóa Payload Phản Hồi Thất Bại
Khi có lỗi, API trả về cấu trúc lỗi chi tiết theo [`common.DefaultError`](file:///home/wsl-admin/projects/terra-clone/server/common/error_response.go) (lưu ý: `message` và `reason` luôn sử dụng Tiếng Anh):
```json
{
  "code": 404,
  "status": "Not Found",
  "request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
  "message": "Requested resource not found",
  "reason": "User with the specified identifier does not exist",
  "details": {
    "identifier": "usr_unknown"
  }
}
```

### 3.4. Chiến Lược Phân Trang Con Trỏ (Cursor-Based Pagination)
- **Cơ chế**: Thay vì sử dụng `LIMIT / OFFSET` gây chậm khi dữ liệu lớn, hệ thống áp dụng con trỏ theo cột `id` kiểu `UUIDv7` kết hợp thời gian phía CSDL nhưng **chỉ để lộ `public_id` ra bên ngoài**:
  - Client gửi `cursor=usr_...` hoặc `cursor=cmp_...`
  - Repository tự động phân giải `public_id` thành UUID nội bộ và thực thi:
    $$\text{WHERE id} < \text{cursor\_uuid ORDER BY id DESC LIMIT 20}$$
  - Phản hồi trả về `next_cursor` là `public_id` của bản ghi cuối cùng trong trang.
- Struct [`Paging`](file:///home/wsl-admin/projects/terra-clone/server/common/pagination.go#L3-L7) đóng gói các giá trị `total`, `cursor`, `next_cursor`.

---

## 4. 🗄️ Tiêu Chuẩn Cơ Sở Dữ Liệu & Migrations (Database Standards)

### 4.1. Quy Chuẩn Đặt Tên Bảng & Cột
- Tên bảng: Sử dụng danh từ số nhiều, định dạng `snake_case` (ví dụ: `users`, `user_credentials`, `user_profiles`).
- Khóa chính: Bắt buộc sử dụng kiểu `UUID` có giá trị mặc định là hàm `uuidv7()`:
  ```sql
  id UUID PRIMARY KEY DEFAULT uuidv7()
  ```
- Cột thời gian: Mọi bảng bắt buộc phải có hai cột quản lý dấu thời gian:
  ```sql
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
  ```
- Ràng buộc khóa ngoại: Phải khai báo tường minh hành vi xóa:
  ```sql
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE
  ```

### 4.2. Quy Chuẩn Đánh Chỉ Mục (Indexing Rules)
- B-Tree index: Đặt tên theo cú pháp `idx_{tên_bảng}_{tên_cột}` (ví dụ: `idx_users_email`, `idx_users_public_id`).
- GIN index cho JSONB: Đặt tên theo cú pháp `idx_{tên_bảng}_{tên_cột}` chỉ định rõ thuật toán GIN (ví dụ: `idx_user_profiles_preferences ON user_profiles USING gin (preferences)`).
- Composite index: Sắp xếp cột có tính chọn lọc cao (cardinality) đứng trước (ví dụ: `idx_user_credentials_provider ON user_credentials(auth_provider, provider_user_id)`).

### 4.3. Nguyên Tắc Viết SQL Migration
- Sử dụng công cụ `golang-migrate/v4`.
- Luôn tạo đồng thời cặp tệp: `{thứ_tự}_{tên}.up.sql` và `{thứ_tự}_{tên}.down.sql`.
- Nguyên tắc lũy kế (Idempotent): Luôn dùng cú pháp an toàn:
  - `CREATE TABLE IF NOT EXISTS ...`
  - `DROP TABLE IF EXISTS ... CASCADE;`
  - `CREATE INDEX IF NOT EXISTS ...`
- Thêm chú thích cột bằng lệnh `COMMENT ON COLUMN` để tự động tài liệu hóa lược đồ.

### 4.4. Quy Chuẩn Ánh Xạ Entity Model (GORM & Golang-Migrate)
- Vì toàn bộ cấu trúc bảng, kiểu dữ liệu, giá trị mặc định (`DEFAULT`), ràng buộc toàn vẹn (`NOT NULL`, `UNIQUE`, `REFERENCES`) đều do `golang-migrate` quản lý độc lập qua SQL scripts, các struct Entity trong `server/models/` **chỉ đóng vai trò là Data Mapper** cho GORM.
- **Quy tắc khai báo struct tag `gorm`**:
  - Chỉ định danh cột tương ứng: `gorm:"column:<ten_cot>"`.
  - Khóa chính: Khai báo `primaryKey` khi cột khóa chính không dùng tên mặc định `id` (ví dụ: `UserID uuid.UUID `gorm:"column:user_id;primaryKey"``).
  - Quan hệ: Sử dụng `foreignKey` và `references` để hỗ trợ cơ chế nạp sẵn quan hệ `Preload()`.
  - **CẤM**: Không khai báo các tag DDL dư thừa của GORM AutoMigrate như `type:...`, `not null`, `default:...`, `unique`, `autoCreateTime`, `autoUpdateTime`.

---

## 5. 📜 Tiêu Chuẩn Ghi Log & Bảo Mật (Logging & Security)

### 5.1. Tiêu Chuẩn Structured Logging
- Mọi thông tin ghi nhận lỗi hoặc sự kiện hệ thống phải dùng `log/slog` thay thế cho `fmt.Print` hoặc `log.Fatal`.
- **Thông điệp log (`msg`)**: Bắt buộc viết hoàn toàn bằng **Tiếng Anh (English)**, không chứa tiền tố `[VN]` trong chuỗi message.
- Phân loại cấp độ log:
  - `LevelInfo`: Dùng cho các sự kiện dòng chảy thông thường (khởi động dịch vụ, định tuyến hoàn tất, kết nối CSDL thành công).
  - `LevelWarn`: Dùng cho các tình huống bất thường nhưng ứng dụng tự xử lý được (file `.env` thiếu và dùng fallback biến hệ thống, số lần đăng nhập sai tăng cao).
  - `LevelError`: Dùng cho các lỗi ảnh hưởng đến xử lý request hoặc mất kết nối ngoại vi.

### 5.2. Bảo Mật Thông Tin Nhật Ký (Log Sanitization)
- **Tuyệt đối không ghi vào log**: Mật khẩu dạng plain text, mã hash mật khẩu, Secret Key, Access Token / Refresh Token, số thẻ ngân hàng hoặc thông tin định danh cá nhân nhạy cảm (PII).

---

## 6. 🧪 Tiêu Chuẩn Kiểm Thử Backend (Backend Testing Standards)

### 6.1. Nguyên Tắc Kiểm Thử Trọng Tâm (Focused Testing Scope)
Để tối ưu hóa chi phí phát triển, giảm gánh nặng bảo trì mock và đảm bảo chất lượng phần mềm tập trung vào đúng nơi phát sinh giá trị nghiệp vụ, dự án thiết lập quy tắc bắt buộc:

- **Chỉ viết Unit Test cho tầng Service (`server/services/`)**:
  - Mọi unit test cho backend chỉ được tập trung duy nhất vào các tệp thuộc tầng Service (ví dụ: `auth_service_test.go`, `user_service_test.go`).
  - **Lý do**: Tầng Service chứa 100% logic nghiệp vụ cốt lõi (business rules), các thuật toán tính toán, quy tắc xác thực quyền hạn và điều phối luồng xử lý. Tầng này làm việc hoàn toàn với Go struct thuần túy và `context.Context`, độc lập với HTTP transport và cơ sở dữ liệu thật, giúp unit test chạy nhanh, ổn định và có tính bao phủ cao nhất.
- **CẤM viết test cho các thành phần khác**:
  - **Controllers (`controllers/`)**: Không viết test. Tầng Controller chỉ thực hiện binding DTO từ HTTP request và chuyển tiếp lời gọi sang Service; việc viết test cho Controller gây lãng phí thời gian và trùng lặp với Service test.
  - **Repositories (`repositories/`)**: Không viết test. Tầng Repository đóng gói các câu lệnh truy vấn qua GORM/SQL; không viết unit test với mock ORM phức tạp hoặc in-memory DB dễ gây sai lệch hành vi so với PostgreSQL thực tế.
  - **Models, Middleware, Common, Internal Config/Router, Migrations**: Không viết test, trừ khi có yêu cầu đặc thù bằng văn bản từ Lead.

### 6.2. Quy Chuẩn Hiện Thực Unit Test Cho Service
1. **Vị trí & Đặt tên tệp (Thư mục `server/tests/` riêng biệt)**:
   - Toàn bộ các tệp test và mock helper của tầng Service được tách biệt hoàn toàn sang thư mục [`server/tests/`](file:///home/wsl-admin/projects/terra-clone/server/tests):
     - `server/tests/auth_service_test.go`
     - `server/tests/user_service_test.go`
     - `server/tests/mock_test.go`
   - Khai báo `package tests` nhằm kiểm thử độc lập theo chuẩn Black-Box Testing thông qua các public interface của Service (`services.AuthService`, `services.UserService`), giữ cho thư mục thực thi `server/services/` luôn tinh gọn và thuần túy code nghiệp vụ.
2. **Kỹ thuật Mocking**:
   - Khởi tạo Mock struct hiện thực Interface của tầng Repository (hoặc các dịch vụ bên thứ ba) mà Service phụ thuộc.
   - Kiểm thử độc lập logic của Service mà không cần bất kỳ kết nối mạng nội bộ hoặc CSDL nào.
3. **Quy chuẩn Ngôn ngữ & Chú thích trong Test**:
   - **Chú thích (Comments)**: Bắt buộc có tiền tố `[VN]` và viết bằng tiếng Việt rõ nghĩa (ví dụ: `// [VN] Kiểm thử trường hợp đăng nhập thành công với mật khẩu chính xác`).
   - **Thông điệp (Messages & Assertions)**: Toàn bộ thông điệp test runner, failure reason trong `assert` hoặc `require` (thư viện `testify`) phải viết 100% bằng **Tiếng Anh (English)**.
4. **Lệnh thực thi kiểm thử chuẩn**:
   ```bash
   # Chạy toàn bộ test trong server/tests/ có kiểm tra race detector và đo coverage tầng services
   cd server && go test -v -coverpkg=./services/... -race ./tests/...
   # Hoặc dùng lệnh Make từ root
   make test
   ```
