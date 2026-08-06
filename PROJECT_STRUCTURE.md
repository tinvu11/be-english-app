# Cấu trúc dự án Go Clean Template

Tài liệu này mô tả cấu trúc **hiện tại** của dự án sau khi loại bỏ các domain `content` và `translation`.

## 1. Tổng quan kiến trúc

Dự án tổ chức theo Clean Architecture với luồng phụ thuộc chính:

```text
HTTP request
    ↓
Router / Middleware
    ↓
Controller
    ↓
Use case
    ↓
Repository interface
    ↓
PostgreSQL repository
    ↓
PostgreSQL
```

Nguyên tắc quan trọng:

- Controller chỉ xử lý HTTP, validation và chuyển đổi request/response.
- Use case chứa nghiệp vụ và không phụ thuộc trực tiếp vào Fiber hoặc PostgreSQL.
- Repository chịu trách nhiệm truy xuất dữ liệu.
- Entity là cấu trúc dữ liệu nghiệp vụ dùng chung giữa các layer.
- Dependency injection được thực hiện thủ công tại `internal/app/app.go`.

## 2. Cây thư mục

```text
.
├── cmd/
│   └── app/
│       └── main.go
├── config/
│   └── config.go
├── docs/
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── integration-test/
│   ├── Dockerfile
│   ├── helpers_test.go
│   └── user_test.go
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   └── migrate.go
│   ├── controller/
│   │   └── restapi/
│   │       ├── middleware/
│   │       ├── v1/
│   │       │   ├── admin/
│   │       │   ├── request/
│   │       │   └── response/
│   │       └── router.go
│   ├── entity/
│   ├── repo/
│   │   └── persistent/
│   │       └── user/
│   └── usecase/
│       └── user/
├── migrations/
├── nginx/
├── pkg/
│   ├── firebaseauth/
│   ├── httpserver/
│   ├── jwt/
│   ├── logger/
│   ├── postgres/
│   └── tracing/
├── docker-compose.yml
├── docker-compose-integration-test.yml
├── Dockerfile
├── go.mod
├── Makefile
└── README.md
```

## 3. Điểm khởi động ứng dụng

### `cmd/app/main.go`

Đây là entry point của chương trình:

1. Đọc biến môi trường bằng `config.NewConfig()`.
2. Gọi `app.Run(cfg)` để khởi tạo và chạy ứng dụng.

File này nên được giữ nhỏ và không chứa nghiệp vụ.

### `internal/app/app.go`

Đây là composition root, chịu trách nhiệm nối các thành phần:

```text
PostgreSQL
  → User repository
  → User use case
  → REST controller
```

File này thực hiện:

- Khởi tạo logger.
- Khởi tạo OpenTelemetry.
- Kết nối PostgreSQL.
- Khởi tạo Firebase token verifier.
- Khởi tạo repository và use case.
- Khởi tạo HTTP server và router.
- Xử lý graceful shutdown khi nhận `SIGTERM` hoặc interrupt.

Khi thêm một domain mới, constructor của repository và use case thường được gọi tại đây.

### `internal/app/migrate.go`

File chỉ được compile khi dùng build tag `migrate`:

```bash
go run -tags migrate ./cmd/app
```

Nó đọc `PG_URL`, kết nối PostgreSQL và chạy toàn bộ migration chưa được áp dụng trước khi application khởi động.

## 4. Configuration

### `config/config.go`

Cấu hình được đọc từ biến môi trường bằng thư viện `caarlos0/env`.

Các nhóm cấu hình hiện có:

| Nhóm | Mục đích |
|---|---|
| `App` | Tên và phiên bản service |
| `HTTP` | Cổng HTTP và Fiber prefork |
| `Log` | Mức log |
| `PG` | PostgreSQL URL và pool size |
| `Firebase` | Project ID và service-account file |
| `Metrics` | Bật/tắt Prometheus metrics |
| `Swagger` | Bật/tắt Swagger UI |
| `Tracing` | OpenTelemetry endpoint và sampling |

Các giá trị local thường được đặt trong `.env`, còn Docker Compose truyền chúng qua `environment`.

## 5. HTTP server và router

### `pkg/httpserver/`

Bao bọc Fiber server và cung cấp:

- Tùy chọn port.
- Prefork mode.
- Start, shutdown và notification channel.

### `internal/controller/restapi/router.go`

Router cấp ứng dụng cấu hình:

- Request logger.
- Panic recovery.
- CORS.
- Prometheus tại `/metrics`.
- Swagger tại `/swagger/*` khi được bật.
- Health check tại `/healthz`.
- API group `/v1`.
- OpenTelemetry middleware khi tracing được bật.

### `internal/controller/restapi/v1/router.go`

Các API hiện tại:

| Method | Endpoint | Quyền |
|---|---|---|
| `GET` | `/v1/user/profile` | User đã đăng nhập |
| `POST` | `/v1/admin/auth/login` | Firebase token của admin |
| `GET` | `/v1/admin/me` | Chỉ admin |
| `GET` | `/v1/admin/dashboard` | Chỉ admin |

Route user sử dụng middleware `Auth`. Admin login tự xác minh Firebase ID token từ request body; `/admin/me` và dashboard sử dụng lần lượt `Auth` và `AdminOnly`.

### `internal/controller/restapi/v1/admin/`

Module controller dành riêng cho quản trị viên:

- `router.go`: đăng ký public/protected admin routes.
- `auth.go`: đăng nhập bằng Firebase ID token và lấy admin hiện tại.
- `dashboard.go`: dashboard endpoint.
- `controller.go`: dependencies của module.
- `request/` và `response/`: DTO riêng của admin.

### `internal/controller/restapi/v1/controller.go`

Struct `V1` giữ dependency dùng bởi controller:

- `usecase.User`
- Logger
- Validator

### Request và response DTO

```text
internal/controller/restapi/v1/request/
internal/controller/restapi/v1/response/
```

- Request DTO mô tả dữ liệu client gửi lên và validation rule.
- Response DTO mô tả response HTTP khi entity không phù hợp để trả trực tiếp.
- Không đặt SQL hoặc nghiệp vụ trong DTO.

## 6. Authentication và authorization

### `pkg/firebaseauth/firebase.go`

Khởi tạo Firebase Admin SDK và xác minh Firebase ID token.

Client gửi token theo header:

```http
Authorization: Bearer <firebase-id-token>
```

### `internal/controller/restapi/middleware/auth.go`

Middleware `Auth` thực hiện:

1. Đọc Bearer token.
2. Xác minh token với Firebase.
3. Gọi `User.Authenticate` để tìm hoặc tạo local user.
4. Đặt `user`, `userID` và `firebaseUID` vào request context.
5. Cho request đi tiếp.

Middleware `RequireRole(role)` kiểm tra role của local user. `AdminOnly()` là shortcut cho:

```go
RequireRole(entity.RoleAdmin)
```

Hai role hiện tại:

```text
user
admin
```

## 7. Entity layer

### `internal/entity/`

Entity không phụ thuộc Fiber hoặc PostgreSQL implementation.

Các file chính:

| File | Nội dung |
|---|---|
| `user.go` | User entity và role constants |
| `auth.go` | Identity nhận từ Firebase |
| `errors.go` | Lỗi nghiệp vụ dùng chung |

`entity.User` hiện gồm:

- Local UUID.
- Firebase UID.
- Username.
- Email.
- Role.
- Thời gian tạo và cập nhật.

## 8. Use-case layer

### `internal/usecase/contracts.go`

Khai báo interface mà controller sử dụng. Interface `User` hiện cung cấp:

- `Authenticate`
- `Register`
- `Login`
- `GetUser`

`Register` và `Login` thuộc cơ chế local auth cũ và hiện trả `ErrLocalAuthDisabled`. Luồng chính của dự án là Firebase Authentication.

### `internal/usecase/user/user.go`

Chứa nghiệp vụ user:

- Tìm user theo Firebase UID.
- Tạo local user trong lần đăng nhập Firebase đầu tiên.
- Lấy profile user.

### `internal/usecase/user/tracing.go`

Decorator OpenTelemetry cho use case. Nó tạo span nhưng không thay đổi nghiệp vụ.

### Mocks

```text
internal/usecase/mocks_repo_test.go
internal/usecase/mocks_usecase_test.go
```

Hai file được sinh tự động bằng Mockgen. Không nên sửa thủ công.

Tái tạo mocks bằng:

```bash
make mock
```

## 9. Repository layer

### `internal/repo/contracts.go`

Khai báo `UserRepo`, gồm các thao tác:

- Lưu user.
- Tìm theo local ID.
- Tìm theo email.
- Tìm theo Firebase UID.

Use case chỉ biết interface này, không biết chi tiết pgx hoặc SQL.

### `internal/repo/persistent/user/user.go`

Implementation PostgreSQL của `UserRepo`:

- Xây SQL bằng Squirrel.
- Thực thi query qua pgx pool.
- Scan record thành `entity.User`.
- Chuyển lỗi PostgreSQL thành lỗi nghiệp vụ.

### `internal/repo/persistent/user/tracing.go`

Decorator thêm tracing span cho các repository operation.

## 10. Database và migrations

### PostgreSQL

Kết nối được quản lý trong `pkg/postgres/` bằng pgx pool.

### Migration hiện tại

| Migration | Chức năng |
|---|---|
| `20260403000001_create_users` | Tạo bảng `users` |
| `20260403000002_create_tasks` | Tạo bảng `tasks` |
| `20260725000001_add_firebase_auth` | Thêm `firebase_uid`, cho phép password null |
| `20260727000002_add_role_to_users` | Thêm role `user/admin` |
| `20260806000001_create_learning_schema` | Tạo schema ngôn ngữ, nội dung học, caption và tiến độ |

Schema ứng dụng hiện có các bảng nghiệp vụ:

```text
users
tasks
languages
levels
level_translations
topics
topic_translations
channels
videos
video_topics
video_captions
caption_translations
user_watch_later
user_watch_history
dictation_progress
shadowing_attempts
```

Ngoài ra, Golang Migrate tự quản lý bảng `schema_migrations`.

Lưu ý:

- Hiện không còn code entity/repository/controller cho `tasks`, nhưng migration tạo bảng vẫn còn.
- Các bảng learning schema mới chỉ có ở database layer; chưa có entity, repository, use case hoặc API tương ứng.
- `users.id` tiếp tục dùng UUID để tương thích với Firebase và code hiện tại. Vì vậy các khóa ngoại `user_id` cũng dùng UUID.
- `native_language_id` và `target_language_id` nullable để lần đăng nhập Firebase đầu tiên vẫn tạo được user trước bước onboarding.

### Quy tắc migration

- Mỗi thay đổi schema có một file `.up.sql` và `.down.sql`.
- Không sửa migration đã chạy trên môi trường dùng chung.
- Với database development có thể reset volume và chạy lại từ đầu.
- Với production phải tạo migration mới để thay đổi hoặc xóa schema.

## 11. Swagger

Swagger annotation nằm gần handler trong controller. Các file generated nằm trong `docs/`:

```text
docs/docs.go
docs/swagger.json
docs/swagger.yaml
```

Tái tạo Swagger:

```bash
make swag-v1
```

Không nên sửa trực tiếp các file generated vì lần generate tiếp theo sẽ ghi đè chúng.

## 12. Logging, metrics và tracing

### Logging

`pkg/logger/` cung cấp structured logging bằng Zerolog. Request logger middleware ghi lại request HTTP.

### Metrics

Khi `METRICS_ENABLED=true`, Prometheus metrics được công bố tại:

```text
GET /metrics
```

### Tracing

`pkg/tracing/` cấu hình OpenTelemetry và OTLP exporter. Khi tracing được bật, request, use case và repository có thể nằm trong cùng một trace.

Docker Compose sử dụng Jaeger làm backend quan sát trace.

## 13. Tests

### Unit tests

Các unit test hiện có:

- Auth/admin middleware.
- User use case.
- JWT package cũ.
- Logger.

Chạy unit test chính:

```bash
make test
```

### Integration tests

Nằm trong `integration-test/` và chạy qua Docker Compose:

```bash
make compose-up-integration-test
```

Lưu ý: integration test user hiện vẫn mô tả `/auth/register` và `/auth/login` của local authentication cũ, trong khi runtime hiện sử dụng Firebase và đã vô hiệu hóa local auth. Bộ test này cần được cập nhật hoặc loại bỏ.

## 14. Docker

### `docker-compose.yml`

Các service chính:

| Service | Mục đích |
|---|---|
| `db` | PostgreSQL |
| `app` | Go application |
| `jaeger` | OpenTelemetry trace UI |
| `nginx` | Reverse proxy |

Database sử dụng named volume `db_data`. Dữ liệu vẫn tồn tại sau khi container bị xóa.

Reset database development:

```bash
docker compose down -v
docker compose up --build -d db
make run
```

Lệnh `down -v` xóa toàn bộ dữ liệu database, chỉ dùng khi chắc chắn không cần dữ liệu cũ.

## 15. Cách thêm một API/domain mới

Ví dụ thêm domain `product`:

1. Tạo migration cho bảng `products`.
2. Tạo `entity.Product`.
3. Thêm `ProductRepo` vào `internal/repo/contracts.go`.
4. Implement repository tại `internal/repo/persistent/product/`.
5. Thêm `Product` interface vào `internal/usecase/contracts.go`.
6. Implement nghiệp vụ tại `internal/usecase/product/`.
7. Tạo request/response DTO nếu cần.
8. Viết controller tại `internal/controller/restapi/v1/product.go`.
9. Đăng ký route trong `v1/router.go`.
10. Nối repository/use case trong `internal/app/app.go`.
11. Generate mocks và Swagger.
12. Chạy format và test.

Các lệnh thường dùng:

```bash
make mock
make swag-v1
make format
make test
```

## 16. Quy tắc phụ thuộc

Hướng phụ thuộc mong muốn:

```text
controller → usecase interface → entity
usecase implementation → repository interface → entity
repository implementation → entity + PostgreSQL
app → tất cả constructor để ghép dependency
```

Không nên:

- Gọi SQL trực tiếp trong controller.
- Trả `fiber.Ctx` xuống use case.
- Đưa validation HTTP vào repository.
- Import controller từ entity/use case/repository.
- Đặt nghiệp vụ trong migration hoặc DTO.

## 17. Các điểm cần dọn tiếp

Trạng thái hiện tại còn một số thành phần legacy:

- Migration `tasks` còn tồn tại nhưng không có code domain task.
- `pkg/jwt` không còn nằm trong luồng Firebase chính.
- `Register` và `Login` vẫn còn trong interface nhưng đã bị vô hiệu hóa.
- Integration test vẫn kiểm tra local register/login cũ.
- README gốc còn mô tả một số server/capability từ upstream template không còn trong runtime hiện tại.

Nên xử lý các điểm này trước khi phát triển domain mới để cấu trúc dự án phản ánh đúng chức năng thực tế.
