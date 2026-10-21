# 📄 Tài Liệu Yêu Cầu Thiết Kế Sản Phẩm (PDR - Product Design Requirement)
## Dự án: Terra Clone

---

## 1. 📌 Bối cảnh & Tầm nhìn Dự án (Context & Vision)

### 1.1. Bối cảnh (Background)
**Terra Clone** là hệ thống nền tảng dịch vụ số và thương mại hiện đại, được xây dựng nhằm cung cấp giải pháp backend hiệu năng cao, linh hoạt và bảo mật. Dự án mô phỏng và kế thừa các ưu điểm kiến trúc của nền tảng Terra, đáp ứng yêu cầu chịu tải cao, sẵn sàng mở rộng quy mô từ quy mô khởi tạo đến môi trường sản xuất phục vụ hàng triệu người dùng.

### 1.2. Tầm nhìn (Vision)
Xây dựng một hệ sinh thái monorepo/microservices hoàn chỉnh với:
- **Cổng giao tiếp tập trung (API Gateway)** xử lý lưu lượng và cân bằng tải.
- **Dịch vụ Backend (Go API Server)** viết bằng Golang với hiệu năng cao, tối ưu tài nguyên và cấu trúc mã nguồn phân tầng rõ ràng (Clean / Layered Architecture).
- **Cơ sở dữ liệu (PostgreSQL 18)** mạnh mẽ, tối ưu hóa lưu trữ với UUIDv7 và JSONB GIN Index.
- **Chuẩn DevOps & Container hóa** tự động, hỗ trợ liền mạch từ môi trường Local/Dev đến Staging và Hardened Production.

---

## 2. 🎯 Mục tiêu Sản phẩm (Product Goals)

### 2.1. Mục tiêu Nghiệp vụ (Business Goals)
1. **Quản lý Định danh & Xác thực Tập trung**: Hỗ trợ xác thực tập trung bằng Email/Password bảo mật cao (chống brute-force, thu hồi session hàng loạt).
2. **Cá nhân hóa Trải nghiệm Người dùng**: Cung cấp khả năng lưu trữ và truy vấn cấu hình cá nhân hóa nhanh chóng thông qua định dạng dữ liệu linh hoạt (JSONB).
3. **Bảo mật Cấp Doanh nghiệp**: Ngăn chặn brute-force tấn công mật khẩu, hỗ trợ cơ chế thu hồi phiên tức thời (Token Invalidation qua `token_version`).

### 2.2. Mục tiêu Kỹ thuật (Technical Goals)
1. **Độ trễ thấp & Thông lượng cao**: Thời gian phản hồi API core dưới 50ms ở mức P95.
2. **Khả năng quan sát toàn diện (Observability)**: Truy vết request end-to-end với Request ID, Access Logging theo thời gian thực và Structured Logging chuẩn JSON qua [`log/slog`](file:///home/wsl-admin/projects/terra-clone/server/internal/logger.go).
3. **Môi trường Phát triển Nhất quán**: Live-reload bằng [Air](file:///home/wsl-admin/projects/terra-clone/server/.air.toml) trên Local Dev và đóng gói image tối giản, bảo mật cao (Non-root `10001:10001`) trên Production.

---

## 3. 👥 Chân dung Người dùng & Stakeholders (User Personas)

| Persona | Vai trò | Nhu cầu chính | Thách thức / Pain Points |
| :--- | :--- | :--- | :--- |
| **End-User (Khách hàng)** | Người dùng cuối đăng ký & sử dụng dịch vụ trên web/app | Đăng nhập mượt mà, bảo mật tài khoản, tốc độ tải trang nhanh, lưu tùy chọn cá nhân | Lo ngại lộ lọt mật khẩu, session đăng nhập chập chờn khi đổi thiết bị |
| **System Admin / Operator** | Kỹ sư vận hành & bảo trì hệ thống | Theo dõi trạng thái sức khỏe dịch vụ, kiểm tra log, phân tích lỗi, sao lưu & khôi phục dữ liệu | Khó debug khi thiếu trace ID, log quá tải gây nghẽn đĩa cứng |
| **Backend Developer** | Kỹ sư phát triển tính năng | Luồng code rõ ràng, phân tầng chuẩn, dễ viết unit test, migration CSDL tự động và an toàn | Đụng độ môi trường, migrate CSDL lỗi giữa Local và Production |

---

## 4. 🧩 Yêu cầu Tính năng (Functional Requirements - FR)

### 4.1. Phân hệ Quản lý Tài khoản & Định danh (Identity & Access Management)
- **FR-01 (Đăng ký tài khoản & Định danh công khai)**:
  - Cho phép người dùng đăng ký tài khoản với email hợp lệ và mật khẩu bảo mật.
  - Mỗi tài khoản sở hữu một mã `public_id` định danh duy nhất công khai cho người dùng và khóa chính nội bộ `UUIDv7` (tự động sắp xếp theo thời gian thực) theo lược đồ bảng [`users`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000001_users_table.up.sql#L1-L10).
- **FR-02 (Đăng nhập & Chống dò quét brute-force)**:
  - Xác thực người dùng bằng email và mật khẩu.
  - Ghi nhận `failed_login_attempts`. Nếu vượt quá ngưỡng cho phép, tự động khóa tài khoản tạm thời qua trường `locked_until` tại bảng [`user_credentials`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000002_user_credentials_table.up.sql#L8-L10).
- **FR-03 (Social Login / OAuth - Kế hoạch giai đoạn sau)**:
  - Dự kiến mở rộng liên kết tài khoản với bên thứ ba (Google, Apple, Facebook,...); trong giai đoạn hiện tại tập trung cốt lõi vào xác thực Email & Password an toàn và chưa triển khai các trường OAuth.
- **FR-04 (Thu hồi phiên & Đổi mật khẩu an toàn)**:
  - Quản lý phiên qua `token_version`. Khi người dùng đổi mật khẩu hoặc thực hiện "Đăng xuất khỏi tất cả thiết bị", hệ thống tăng giá trị `token_version` để lập tức vô hiệu hóa toàn bộ JWT token cũ.

### 4.2. Phân hệ Hồ sơ & Cá nhân hóa (User Profiles & Preferences)
- **FR-05 (Hồ sơ người dùng)**:
  - Quản lý ngày sinh, giới tính, số điện thoại, địa chỉ theo bảng [`user_profiles`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000003_user_profiles_table.up.sql#L1-L10).
  - Khóa ngoại `user_id` liên kết `ON DELETE CASCADE` với bảng `users`.
- **FR-06 (Tùy chọn linh hoạt - JSONB Preferences)**:
  - Lưu trữ theme, notification settings, language preferences dưới cột `preferences JSONB`.
  - Hỗ trợ truy vấn JSON tốc độ cao bằng chỉ mục GIN (`idx_user_profiles_preferences`).

### 4.3. Phân hệ Gateway & Điều hướng Lưu lượng (Gateway & Routing)
- **FR-07 (Health Check & Giám sát sống còn)**:
  - Cung cấp endpoint [`GET /healthz`](file:///home/wsl-admin/projects/terra-clone/nginx/conf.d/default.conf#L10-L14) trên Nginx trả về `200 OK` để các load balancer kiểm tra trạng thái container.
  - Cung cấp endpoint [`GET /ping`](file:///home/wsl-admin/projects/terra-clone/server/internal/router.go#L19-L23) (hoặc `/api/ping` qua Gateway) để kiểm tra trạng thái hoạt động của Go Server Engine.
- **FR-08 (Điều hướng API & Frontend Proxy)**:
  - Mọi yêu cầu `/api/*` được chuyển tiếp nguyên vẹn tới Go API Server (`http://server:8080`).
  - Mọi yêu cầu `/*` còn lại được chuyển tiếp tới Frontend Next.js (`http://frontend:3000`), hỗ trợ WebSocket Upgrade cho Hot Module Replacement (HMR).

---

## 5. 🛡️ Yêu cầu Phi chức năng (Non-Functional Requirements - NFR)

### 5.1. Hiệu năng & Khả năng chịu tải (Performance & Scalability)
- **NFR-01 (Throughput & Latency)**: Go server xây dựng trên framework Gin Gonic siêu nhẹ kết hợp bộ đệm kết nối Connection Pool của PostgreSQL, cho phép xử lý đồng thời hàng nghìn RPS với mức tiêu thụ RAM dưới 50MB.
- **NFR-02 (Database Indexing)**: Toàn bộ các trường định danh và tìm kiếm chính (`users.public_id`, `users.email`, `preferences`) đều được đánh chỉ mục B-Tree và GIN tối ưu.

### 5.2. Bảo mật & Toàn vẹn Dữ liệu (Security)
- **NFR-03 (Kiến trúc Hardened Container)**: Image Docker cho Production được tối giản dựa trên Alpine Linux 3.24, loại bỏ toàn bộ compiler/build tools, chạy dưới tài khoản non-root (`UID/GID 10001:10001`), không mở cổng CSDL ra internet.
- **NFR-04 (Bảo vệ Gateway)**: Nginx được trang bị các HTTP Security Headers chuẩn: `X-Frame-Options: SAMEORIGIN`, `X-Content-Type-Options: nosniff`, `X-XSS-Protection`, `Referrer-Policy: strict-origin-when-cross-origin`.
- **NFR-05 (Bảo vệ dữ liệu nhạy cảm)**: Cấu trúc phản hồi lỗi chuẩn [`DefaultError`](file:///home/wsl-admin/projects/terra-clone/server/common/error_response.go#L12-L53) tách biệt rõ ràng giữa thông báo trả về cho client (`message`, `reason`) và thông tin debug nội bộ (`debug`, `StackTrace`), tuyệt đối không để rò rỉ cấu trúc SQL hoặc stacktrace ra bên ngoài.

### 5.3. Khả năng bảo trì & Vận hành (Maintainability & Observability)
- **NFR-06 (Structured Logging)**: Định dạng log chuẩn JSON với `slog`, ghi nhận đầy đủ thông tin: IP, Latency, Method, Path, User-Agent, Status Code.
- **NFR-07 (Tự động xoay vòng log)**: Sử dụng thư viện `lumberjack/v3` với giới hạn dung lượng 10MB/file, lưu trữ tối đa 5 bản sao lưu trong vòng 28 ngày, tự động nén `.gz` và hỗ trợ flush buffer khi tắt server qua cơ chế `defer cleanupLogger()`.

---

## 6. 🗺️ Ma trận Truy vết Yêu cầu (Traceability Matrix)

| Mã Yêu cầu | Mô tả tính năng | Thành phần phụ trách | File hiện thực / Schema |
| :--- | :--- | :--- | :--- |
| **FR-01** | Tạo tài khoản, UUIDv7 ID | Go Server / Postgres | [`000001_users_table.up.sql`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000001_users_table.up.sql) |
| **FR-02** | Chống brute-force đăng nhập | Go Auth Service / Postgres | [`000002_user_credentials_table.up.sql`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000002_user_credentials_table.up.sql) |
| **FR-03** | Hỗ trợ OAuth Social Login (Kế hoạch) | Go Auth Service / Postgres | Dự kiến giai đoạn sau |
| **FR-04** | Thu hồi token qua version | Go Middleware & Auth | [`000002_user_credentials_table.up.sql`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000002_user_credentials_table.up.sql) |
| **FR-05, 06** | Hồ sơ người dùng & JSONB Preferences | Go Server / Postgres | [`000003_user_profiles_table.up.sql`](file:///home/wsl-admin/projects/terra-clone/server/migrations/000003_user_profiles_table.up.sql) |
| **FR-07, 08** | Reverse proxy, Healthcheck & HMR | Nginx Gateway | [`nginx.conf`](file:///home/wsl-admin/projects/terra-clone/nginx/nginx.conf), [`default.conf`](file:///home/wsl-admin/projects/terra-clone/nginx/conf.d/default.conf) |
| **NFR-03** | Bảo mật Container Production | Dockerfile Production | [`Dockerfile.product`](file:///home/wsl-admin/projects/terra-clone/server/Dockerfile.product) |
| **NFR-06, 07** | Structured Logging & Xoay log | Logger Middleware | [`logger.go`](file:///home/wsl-admin/projects/terra-clone/server/internal/logger.go), [`middleware/logger.go`](file:///home/wsl-admin/projects/terra-clone/server/middleware/logger.go) |
