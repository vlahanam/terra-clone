# 🗺️ Lộ Trình Phát Triển Dự Án (Project Roadmap)
## Dự án: Terra Clone

---

## 1. 🧭 Tổng Quan Lộ Trình (Roadmap Overview)

Lộ trình phát triển của dự án **Terra Clone** được chia thành 5 giai đoạn chiến lược, hướng tới việc xây dựng một nền tảng dịch vụ số hiện đại, hiệu năng cao và có khả năng chịu tải hàng triệu người dùng.

```mermaid
flowchart TD
    Phase1["<b>Giai Đoạn 1: Nền Tảng Cốt Lõi & Hạ Tầng</b><br/>(Core Infrastructure & Foundation)<br/><i>Trạng thái: ĐÃ HOÀN THÀNH</i>"]
    Phase2["<b>Giai Đoạn 2: Xác Thực & Quản Trị Định Danh</b><br/>(Authentication & IAM)<br/><i>Trạng thái: ĐÃ HOÀN THÀNH</i>"]
    Phase3["<b>Giai Đoạn 3: Hồ Sơ Cá Nhân & Preferences</b><br/>(User Profiles & Dynamic JSONB)<br/><i>Trạng thái: ĐÃ HOÀN THÀNH</i>"]
    Phase4["<b>Giai Đoạn 4: Tích Hợp Frontend Next.js</b><br/>(Web Application & Client Portal)<br/><i>Trạng thái: KẾ HOẠCH TIẾP THEO</i>"]
    Phase5["<b>Giai Đoạn 5: CI/CD, Giám Sát & Hardening</b><br/>(Production Hardening & Observability)<br/><i>Trạng thái: DỰ KIẾN</i>"]

    Phase1 --> Phase2
    Phase2 --> Phase3
    Phase3 --> Phase4
    Phase4 --> Phase5
```

---

## 2. 📅 Chi Tiết Từng Giai Đoạn (Milestones Breakdown)

### 🟢 Giai Đoạn 1: Nền Tảng Hạ Tầng & Khung Kiến Trúc (Foundation & Core Infrastructure)
> **Trạng thái**: ✅ **ĐÃ HOÀN THÀNH (Baseline Achieved)**

- [x] **Cổng Dịch Vụ Gateway (Nginx)**:
  - Thiết lập reverse proxy với Docker internal DNS resolver (`127.0.0.11`).
  - Cấu hình điều hướng `/healthz`, `/api/*`, và `/*` kèm WebSocket Upgrade cho HMR.
  - Tối ưu gzip level 6, buffer size và security headers cho Production.
- [x] **Khung Backend Server (Golang 1.26+ & Gin Gonic)**:
  - Kiến trúc phân tầng Layered Clean Architecture ([`cmd/`](file:///home/wsl-admin/projects/terra-clone/server/cmd), [`internal/`](file:///home/wsl-admin/projects/terra-clone/server/internal), [`controllers/`](file:///home/wsl-admin/projects/terra-clone/server/controllers), [`services/`](file:///home/wsl-admin/projects/terra-clone/server/services), [`repositories/`](file:///home/wsl-admin/projects/terra-clone/server/repositories)).
  - Cơ chế nạp cấu hình và nguyên tắc thất bại sớm (Fail-Fast) trong [`config.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/config.go).
  - Tích hợp Structured Logging JSON chuẩn `slog` và xoay vòng file tự động bằng `lumberjack/v3` với cơ chế bảo toàn log (`defer cleanupLogger()`).
  - Middleware ghi Access Log chi tiết theo thời gian thực tại [`middleware/logger.go`](file:///home/wsl-admin/projects/terra-clone/server/middleware/logger.go).
- [x] **Cơ Sở Dữ Liệu & Migrations**:
  - Khởi tạo PostgreSQL 18 trên nền tảng Alpine Linux siêu nhẹ.
  - Thiết kế và phiên bản hóa lược đồ ban đầu bằng `golang-migrate`:
    - [`000001_users_table`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000001_users_table.up.sql): Khóa chính UUIDv7, unique email, public_id định danh hiển thị công khai.
    - [`000002_user_credentials_table`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000002_user_credentials_table.up.sql): Hỗ trợ đăng nhập email và password, quản lý session token và phòng thủ brute-force.
    - [`000003_user_profiles_table`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000003_user_profiles_table.up.sql): Cột `preferences JSONB` đánh chỉ mục GIN.
- [x] **Chuẩn Hóa Giao Tiếp API**:
  - Chuẩn phản hồi lỗi phân cấp [`common.DefaultError`](file:///home/wsl-admin/projects/terra-clone/server/common/error_response.go) có StackTrace và Request ID.
  - Chuẩn phản hồi thành công [`common.SuccessResponse`](file:///home/wsl-admin/projects/terra-clone/server/common/success_response.go).
  - Phân trang con trỏ hiệu năng cao [`common.Paging`](file:///home/wsl-admin/projects/terra-clone/server/common/pagination.go).
- [x] **DevOps & Môi Trường**:
  - Thiết lập độc lập 3 file Compose: [`docker-compose.yml`](file:///home/wsl-admin/projects/terra-clone/docker-compose.yml) (Dev với Air live-reload), [`docker-compose.staging.yml`](file:///home/wsl-admin/projects/terra-clone/docker-compose.staging.yml) (Staging), [`docker-compose.product.yml`](file:///home/wsl-admin/projects/terra-clone/docker-compose.product.yml) (Production non-root 10001:10001).
  - Tự động hóa toàn diện qua [`Makefile`](file:///home/wsl-admin/projects/terra-clone/Makefile).

---

### 🟢 Giai Đoạn 2: Phân Hệ Xác Thực & Định Danh Người Dùng (Authentication & IAM)
> **Trạng thái**: ✅ **ĐÃ HOÀN THÀNH (Baseline Achieved)**

- [x] **Data Access Layer (Repositories)**:
  - Hiện thực `UserRepository` (Tìm kiếm theo ID/Email/PublicID, tạo mới, phân trang Cursor, cập nhật, xóa).
  - Hiện thực `CredentialRepository` (Lưu hash mật khẩu, cập nhật số lần thử sai, khóa tạm thời tài khoản, cập nhật `token_version`).
  - Hiện thực `ProfileRepository` (Lưu hồ sơ và preferences JSONB với upsert on conflict).
- [x] **Business Logic Layer (Services)**:
  - Hiện thực hàm băm mật khẩu bảo mật sử dụng thuật toán `Bcrypt` với salt ngẫu nhiên.
  - Logic đăng ký tài khoản (Register Flow) kết hợp transaction đảm bảo tạo đồng bộ 3 bản ghi `users`, `user_credentials` và `user_profiles`.
  - Logic đăng nhập (Login Flow) kiểm tra số lần thử sai liên tiếp (`failed_login_attempts`), tự động kích hoạt `locked_until` nếu vượt quá 5 lần thử.
  - Cấp phát cặp JWT token: Access Token (thời hạn 15 phút) và Refresh Token (thời hạn 7 ngày).
  - Logic Refresh Token đối chiếu `token_version` và cấp cặp token mới.
- [x] **HTTP Transport Layer (Controllers & Middleware)**:
  - Khởi tạo controller `AuthController`:
    - `POST /api/v1/auth/register`: Đăng ký tài khoản mới.
    - `POST /api/v1/auth/login`: Xác thực và trả JWT token.
    - `POST /api/v1/auth/refresh`: Cấp mới Access Token bằng Refresh Token.
    - `GET /api/v1/auth/me`: Xem thông tin tài khoản hiện tại.
    - `POST /api/v1/auth/logout-all`: Tăng `token_version` để đăng xuất khỏi toàn bộ thiết bị.
  - Middleware `RequireAuth`: Trích xuất Header `Authorization: Bearer <token>`, giải mã claims và xác thực với `token_version` trong CSDL.
  - Middleware `RequestID`: Tự động gán `X-Request-ID` độc nhất theo chuẩn UUIDv7.
- [x] **Kiểm Thử Đơn Vị Tầng Service (Backend Service Unit Tests)**:
  - Triển khai toàn bộ unit test cho `AuthService` và `UserService` đạt **81.6% statement coverage** độc lập bằng Mocking và `go-sqlmock`.
- [ ] **Tích Hợp OAuth 2.0 (Google Login)**:
  - Kế hoạch giai đoạn sau.

---

### 🟢 Giai Đoạn 3: Quản Lý Hồ Sơ & Cấu Hình Cá Nhân Hóa (User Profiles & Preferences)
> **Trạng thái**: ✅ **ĐÃ HOÀN THÀNH (Baseline Achieved)**

- [x] **CRUD Người Dùng & Hồ Sơ (User CRUD APIs)**:
  - `POST /api/v1/users`: Tạo người dùng mới qua API quản trị kèm mật khẩu và hồ sơ chi tiết.
  - `GET /api/v1/users/:id`: Trích xuất thông tin chi tiết bằng ID (UUIDv7) hoặc `public_id`.
  - `GET /api/v1/users`: Danh sách người dùng hỗ trợ phân trang Cursor hiệu năng cao (`id < cursor ORDER BY id DESC LIMIT 20`), lọc theo trạng thái (`status`) và tìm kiếm (`search`).
  - `PATCH /api/v1/users/:id` & `PUT /api/v1/users/:id`: Cập nhật thông tin cơ bản và thông tin hồ sơ/preferences.
  - `DELETE /api/v1/users/:id`: Xóa người dùng (hỗ trợ soft-delete `status = 'deleted'` và hard-delete).
- [x] **Khai Thác Cấu Hình Động (JSONB Preferences)**:
  - Lưu trữ và cập nhật cấu hình cá nhân hóa tùy biến (`preferences`) trong PostgreSQL JSONB.
  - Tối ưu hóa truy vấn JSON tốc độ cao bằng chỉ mục GIN (`idx_user_profiles_preferences`).
- [x] **Quản Trị Doanh Nghiệp & Tổ Chức (Company Management APIs)**:
  - `POST /api/v1/companies`: Tạo công ty mới (sinh định danh `public_id` tiền tố `cmp_`, kiểm tra tính duy nhất của mã `code` và mã số thuế `tax_id`, liên kết người dùng tạo làm `owner_id`).
  - `GET /api/v1/companies/:id`: Trích xuất thông tin chi tiết công ty theo ID (UUIDv7), `public_id` hoặc mã `code`.
  - `GET /api/v1/companies`: Danh sách công ty hỗ trợ phân trang Cursor (`UUIDv7`), lọc trạng thái (`status`), lọc theo chủ sở hữu (`owner_id`) và tìm kiếm đa trường (`search`).
  - `PUT /api/v1/companies/:id`: Cập nhật thông tin doanh nghiệp, thay đổi/xóa chủ sở hữu và cập nhật trường cấu hình linh hoạt `metadata JSONB` (đánh chỉ mục GIN).
  - `DELETE /api/v1/companies/:id`: Xóa công ty với cơ chế xóa mềm mặc định (`status = 'deleted'`) và xóa cứng (`?hard=true`).
  - Hoàn thiện bộ kiểm thử đơn vị độc lập cho `CompanyService` tại `server/tests/company_service_test.go` đạt độ bao phủ toàn diện.
- [x] **Hỗ Trợ Tiêu Chuẩn Đa Quốc Gia (Multi-national & Internationalization)**:
  - Bổ sung trường thông tin chuẩn ISO cho cá nhân (`country_code`, `nationality`, `timezone`, `locale`, `city`, `state_or_province`, `postal_code`).
  - Bổ sung trường thông tin chuẩn ISO cho công ty (`country_code`, `timezone`, `currency`, `legal_name`, `city`, `state_or_province`, `postal_code`).
  - Tối ưu hóa truy vấn lọc danh sách theo quốc gia bằng chỉ mục B-Tree `idx_user_profiles_country_code` và `idx_companies_country_code`.
- [ ] **Tải Lên Ảnh Đại Diện (Avatar Upload Pipeline - Kế hoạch giai đoạn sau)**:
  - Tích hợp giải pháp lưu trữ đối tượng (S3-compatible: AWS S3 hoặc MinIO container).
  - Cung cấp API sinh Presigned URL an toàn để client tải ảnh trực tiếp lên storage.

---

### 🟣 Giai Đoạn 4: Phát Triển Giao Diện Web (Frontend Next.js Integration)
> **Trạng thái**: ⏳ **DỰ KIẾN (Upcoming)**

- [ ] **Khởi Tạo Dự Án Web Frontend**:
  - Khởi tạo dự án Next.js 14+ (App Router, React Server Components, TypeScript, Tailwind CSS, Shadcn UI).
  - Cấu hình kết nối qua Nginx Gateway tại cổng `:8088`.
- [ ] **Hiện Thực Các Luồng Giao Diện Người Dùng**:
  - Trang xác thực: Đăng nhập, Đăng ký, Quên mật khẩu, Nút đăng nhập Google OAuth.
  - Trang quản lý thông tin tài khoản và cài đặt Preferences.
  - Trang Dashboard chính của dịch vụ Terra Clone.
- [ ] **Quản Lý Trạng Thái & Đồng Bộ Dữ Liệu**:
  - Xử lý Cookie HTTP-only an toàn cho Refresh Token.
  - Sử dụng thư viện React Query / SWR quản lý cache API và tự động refresh token ngầm khi Access Token hết hạn.

---

### ⚪ Giai Đoạn 5: CI/CD, Quan Sát & Tối Ưu Triển Khai (Hardening & Observability)
> **Trạng thái**: ⏳ **DỰ KIẾN (Upcoming)**

- [ ] **Quy Trình Tích Hợp & Triển Khai Liên Tục (CI/CD Pipeline)**:
  - Cấu hình GitHub Actions: Tự động chạy `golangci-lint`, kiểm thử đơn vị (`go test ./...`) và kiểm thử tích hợp.
  - Tự động đóng gói và đẩy Docker image lên GitHub Container Registry (GHCR) khi gắn tag release.
- [ ] **Bộ Đệm Dữ Liệu Tốc Độ Cao (Redis Cache)**:
  - Bổ sung service Redis vào Docker Compose phục vụ lưu trữ blacklist token tạm thời và Rate Limiting chống spam request.
- [ ] **Hệ Thống Giám Sát & Cảnh Báo (Metrics & Monitoring)**:
  - Tích hợp Prometheus client thu thập metrics hiệu năng (RPS, Latency, Số lượng kết nối DB pool).
  - Trực quan hóa dashboard giám sát với Grafana.
  - Cấu hình công cụ Distributed Tracing (OpenTelemetry / Jaeger) truy vết request end-to-end từ Nginx qua Go Server đến Database.
