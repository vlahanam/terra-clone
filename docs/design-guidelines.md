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
  - Tag GORM: `gorm:"column:full_name;type:varchar(100);not null"`.

### 2.2. Quy Tắc Xử Lý Lỗi (Error Handling Guidelines)
1. **Không nuốt lỗi**: Tuyệt đối không gán `_ = err` khi gọi các thao tác có khả năng sinh lỗi nghiêm trọng.
2. **Bọc lỗi giữ nguyên StackTrace**:
   - Sử dụng [`common.DefaultError`](file:///home/wsl-admin/projects/terra-clone/server/common/error_response.go) kết hợp phương thức `.WithTrace(err)` để đóng gói ngữ cảnh lỗi gốc từ thư viện `github.com/pkg/errors`.
   - Mẫu xử lý chuẩn:
     ```go
     user, err := repo.FindUserByID(ctx, id)
     if err != nil {
         return nil, common.ErrNotFound.
             WithReason("Người dùng không tồn tại trên hệ thống").
             WithTrace(err)
     }
     ```
3. **Phân tách thông tin nhạy cảm**:
   - Trường `ReasonField` / `ErrorField`: Chứa thông báo an toàn, thân thiện để gửi tới người dùng.
   - Trường `DebugField` và `StackTrace()`: Chỉ ghi vào file log nội bộ hoặc phục vụ debug ở môi trường `local`, không serialize ra payload JSON gửi tới client ở môi trường `production`.

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
    "id": "018f2a1b-9f3b-7f28-8d54-1b945f3408f1",
    "email": "user@example.com",
    "full_name": "Nguyễn Văn A",
    "status": "active"
  },
  "paging": {
    "total": 1250,
    "cursor": "018f2a1b-9f3b-7f28-8d54-1b945f3408f1",
    "next_cursor": "018f2a1c-4b12-7e90-a321-2e841a1239bc"
  },
  "extra": null
}
```

### 3.3. Chuẩn Hóa Payload Phản Hồi Thất Bại
Khi có lỗi, API trả về cấu trúc lỗi chi tiết theo [`common.DefaultError`](file:///home/wsl-admin/projects/terra-clone/server/common/error_response.go):
```json
{
  "id": "ERR_NOT_FOUND",
  "code": 404,
  "status": "Not Found",
  "request": "d7ef54b1-ec15-46e6-bccb-524b82c035e6",
  "message": "Không tìm thấy tài nguyên được yêu cầu",
  "reason": "Người dùng có ID được yêu cầu không tồn tại",
  "details": {
    "field": "id"
  }
}
```

### 3.4. Chiến Lược Phân Trang Con Trỏ (Cursor-Based Pagination)
- **Cơ chế**: Thay vì sử dụng `LIMIT / OFFSET` gây chậm khi dữ liệu lớn, hệ thống áp dụng con trỏ theo cột `id` kiểu `UUIDv7` kết hợp thời gian:
  $$\text{WHERE id} < \text{cursor ORDER BY id DESC LIMIT 20}$$
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
- B-Tree index: Đặt tên theo cú pháp `idx_{tên_bảng}_{tên_cột}` (ví dụ: `idx_users_email`, `idx_user_credentials_username`).
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

---

## 5. 📜 Tiêu Chuẩn Ghi Log & Bảo Mật (Logging & Security)

### 5.1. Tiêu Chuẩn Structured Logging
- Mọi thông tin ghi nhận lỗi hoặc sự kiện hệ thống phải dùng `log/slog` thay thế cho `fmt.Print` hoặc `log.Fatal`.
- Phân loại cấp độ log:
  - `LevelInfo`: Dùng cho các sự kiện dòng chảy thông thường (khởi động dịch vụ, định tuyến hoàn tất, kết nối CSDL thành công).
  - `LevelWarn`: Dùng cho các tình huống bất thường nhưng ứng dụng tự xử lý được (file `.env` thiếu và dùng fallback biến hệ thống, số lần đăng nhập sai tăng cao).
  - `LevelError`: Dùng cho các lỗi ảnh hưởng đến xử lý request hoặc mất kết nối ngoại vi.

### 5.2. Bảo Mật Thông Tin Nhật Ký (Log Sanitization)
- **Tuyệt đối không ghi vào log**: Mật khẩu dạng plain text, mã hash mật khẩu, Secret Key, Access Token / Refresh Token, số thẻ ngân hàng hoặc thông tin định danh cá nhân nhạy cảm (PII).
