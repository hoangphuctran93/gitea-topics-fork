# Changelog - Gitea Forge Bridge

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [1.1.0] - 2026-03-10

### Added
- **Background Sync Engine:** Introduced `forge_bridge_sync` worker pool queue.
- **Migration Notifier Hook:** Registered `services/notify` interceptor to capture `MigrateRepository` events.
- **GitHub Profile Fetcher:** Added logic to call GitHub API (`https://api.github.com/users/`) and parse response into Gitea object.
- **Migration Auto-fill:** Connected Gitea standard migration template (`github.tmpl`) with AES-decrypted user token.
- **Deduplication Mapping:** Added `ForgeUserMapping` to prevent redundant fetching within 24 hours.

### Changed
- **Router Initialization:** Updated `routers/init.go` to inject `forgebridge.Init()` during the core boot sequence.
- **User Models:** Swapped direct `UpdateUser` calls to `UpdateUserCols` to avoid unintended modifications and reduce DB transactions.

---

# Nhật Ký Thay Đổi - Gitea Forge Bridge

Tất cả những thay đổi đáng chú ý của dự án sẽ được ghi nhận tại file này.

Định dạng dựa trên [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [1.1.0] - 2026-03-10

### Mới Thêm (Added)
- **Động cơ Đồng bộ chạy ngầm (Background Sync Engine):** Tạo hàng đợi `forge_bridge_sync` kiểu worker pool để xử lý bất đồng bộ.
- **Hook Thông Báo Di Chuyển (Migration Notifier Hook):** Đăng ký đánh chặn `services/notify` nhằm thu thập sự kiện sau khi Migrate Repo thành công.
- **Trình kéo Dữ Liệu Hồ Sơ GitHub (GitHub Profile Fetcher):** Thêm logic gọi API GitHub (`https://api.github.com/users/`) và giải mã JSON phản hồi vào đối tượng người dùng nội bộ Gitea (Avatar, Bio).
- **Tự động Điền Di Chuyển (Migration Auto-fill):** Kết nối Template tiêu chuẩn của Gitea (`github.tmpl`) với khoen giải mã AES Token cho phép người dùng thấy token điền sẵn khi bấm cấu hình Migration.
- **Lập bản đồ Chống Trùng lặp (Deduplication Mapping):** Thêm bảng Mapping ngăn việc gọi API liên tục cho cùng một tài khoản (giới hạn 24h).

### Đã Thay Đổi (Changed)
- **Khởi tạo Route (Router Initialization):** Chỉnh sửa `routers/init.go` để chạy hàm `forgebridge.Init()` trong chu kỳ Web Boot của Gitea.
- **Model Người Dùng:** Chuyển các hàm `UpdateUser` thành `UpdateUserCols` giảm thiểu câu lệnh Cập nhật Database và thao tác lock không mong muốn.
