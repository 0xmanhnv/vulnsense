# VulnSense Project - Software Requirements Specification (SRS)

## 1. Project Overview

**Project Name**: VulnSense
**Goal**: Xây dựng một hệ thống tình báo lỗ hổng tổng hợp và so khớp với Asset Inventory từ Splunk, nhằm cảnh báo nhanh các nguy cơ bảo mật cho hệ thống.

## 2. Functional Scope

### 2.1 Feed Ingestion Module

* Thu thập thông tin từ các nguồn sau:

  * OpenCVE (qua API)
  * RSS / Blog / GitHub Exploit / Threat Intelligence Feed
* Chuẩn hóa thông tin lỗ hổng về định dạng thống nhất

### 2.2 Vulnerability Store

* Lưu vết thông tin các lỗ hổng đã thu thập
* Cho phép query theo ID, sản phẩm, severity, nguồn

### 2.3 Asset Inventory Interface

* Kết nối với Splunk qua REST API
* Query dữ liệu asset inventory (hostname, ip, phần mềm đang cài)

### 2.4 Vulnerability Matching Engine

* So khớp CPE / product / version với asset trong Splunk
* Hỗ trợ match chệnh xác và fuzzy (keyword)

### 2.5 Alerting Engine

* Gửi alert qua Slack, email hoặc log vào Splunk
* Lọc theo severity, số asset bị ảnh hưởng

### 2.6 Scheduler

* Tự động chạy feed fetching, matching, alerting theo chu kỳ
* Mặc định: 30 phút/lần (configurable)

## 3. Non-Functional Requirements

* Viết bằng Go, theo kiểu modular service
* Dễ dàng mở rộng, test đơn vị từng module
* Lưu vết đầy đủ logs và alert history
* Dễ deploy với Docker / Kubernetes

## 4. Project Structure

```
/vulnsense
|
|├─ cmd/                # main.go entrypoint
|├─ pkg/
|   |├─ feeds/         # Crawl + parse external feeds
|   |├─ parser/        # Normalize feeds
|   |├─ matcher/       # Match asset vs vuln
|   |├─ splunk/        # Splunk query client
|   |├─ alert/         # Slack / Email output
|   └─ store/          # DB or cache access
|
|├─ configs/           # YAML config
|├─ scripts/           # Seed, setup
|├─ internal/          # Business orchestration
|├─ web/               # Optional web UI
|└─ README.md          # Project guide
```

## 5. Unified Vulnerability Format (UVF)

```json
{
  "id": "CVE-2025-12345" | "ZERO-WEBPANEL-2025-001",
  "source": "nvd" | "threatpost",
  "product": "openssl",
  "version": "3.0.7",
  "description": "Buffer overflow in ...",
  "cpe_guess": "cpe:2.3:a:openssl:openssl:3.0.7:*:*:*:*:*:*:*",
  "cvss": 9.8,
  "published": "2025-06-29",
  "references": ["https://..."],
  "exploit_available": true
}
```

## 6. Future Roadmap: Towards ASM & Vulnerability Management Platform (VMP)

### Giai đoạn mở rộng VulnSense → VMP

| Giai đoạn | Mục tiêu                | Thành phần chính                                                   |
| --------- | ----------------------- | ------------------------------------------------------------------ |
| **P1**    | Triage lỗ hổng          | UI web dashboard, phân loại (Open/Accepted/Patched/False Positive) |
| **P2**    | Gán ownership           | Mapping asset → owner/team; assign ticket/jira ID                  |
| **P3**    | Theo dõi xử lý          | SLA tracking, trạng thái, thời gian xử lý trung bình               |
| **P4**    | Exception & Audit trail | Cho phép bỏ qua có lý do và log lại hành động của người dùng       |
| **P5**    | Tích hợp hệ thống       | Push/Pull Jira, ServiceNow, GitHub Issues                          |

### Sơ đồ workflow: `Vuln → Triage → Patch Tracking`

```
    +------------+    +-------------+    +-----------------+    +--------------+
    | New Vuln   | -> | Asset Match | -> | Risk Triage UI  | -> | Tracking SLA |
    +------------+    +-------------+    +-----------------+    +--------------+
                                                     |
                                                     v
                                           +-------------------+
                                           | Patch / Exception |
                                           +-------------------+
```

### Mockup UI dashboard (module VMP)

```
┌───────────────────────────────────────────────┐
│             🔍 Vulnerability Overview         │
├──────────────┬──────────────┬─────────────────┤
│ ID           │ Product      │ Status          │
├──────────────┼──────────────┼─────────────────┤
│ CVE-2025-123 │ openssl 3.0.7│ Open (3 assets) │
│ ZERO-BLA-001 │ nginx 1.24   │ Patched         │
│ CVE-2024-888 │ apache 2.4.6 │ Risk Accepted   │
└──────────────┴──────────────┴─────────────────┘

⮕ Bộ lọc: severity | team | asset group  
⮕ Hành động: gán chủ sở hữu | thêm ghi chú | cập nhật trạng thái
```

## 7. Documentation Plan

* `README.md`: Giới thiệu nhanh + install
* `docs/architecture.md`: Kiến trúc hệ thống
* `docs/development.md`: Chuẩn code, quy trình CI/CD
* `docs/feed_formats.md`: Tỉnh đồng chuẩn hóa input
* `docs/alerting.md`: Format, webhook, Slack setup
* `docs/vmp.md`: Quy trình quản lý lỗ hổng nâng cao (Triage - Ownership - SLA)

---

## Appendix A - `docs/architecture.md`

```
+------------------+       +------------------------+
| External Feeds   |-----> | Feed Ingestion Module  |
| (CVE, Blogs, RSS)|       +------------------------+
                            |
                            v
                  +------------------------+
                  | Feed Parser/Normalizer |
                  +------------------------+
                            |
                            v
                  +------------------------+
                  |  Vulnerability Store   |
                  +------------------------+
                            |
                            v
   +---------------> Matcher Engine <----------------+
   |                        |                         |
   |                        v                         |
+--------+         +------------------+        +-------------+
| Splunk |<------->| Asset Inventory  |        | Alert Engine|
+--------+         +------------------+        +-------------+
                                                   |
                                             +-----------+
                                             | Slack/Email|
                                             +-----------+
```

### Các layer chính:

1. **Feeds Layer**: Thu thập dữ liệu từ nhiều nguồn (OpenCVE, threat blog, GitHub RSS, ExploitDB...).
2. **Parsing Layer**: Normalize tất cả dữ liệu lỗ hổng về dạng thống nhất (UVF).
3. **Storage Layer**: Lưu trữ lỗ hổng theo `id`, `product`, `version`, `cvss`, `references`.
4. **Matching Layer**: So khớp asset từ Splunk với danh sách lỗ hổng dựa vào `product`, `version`, `cpe_guess`.
5. **Alerting Layer**: Gửi cảnh báo qua các kênh Slack, Email, Webhook hoặc log về SIEM.

## Appendix B - `docs/feed_formats.md`

### 1. Yêu cầu định dạng chuẩn sau khi chuẩn hóa:

```json
{
  "id": "CVE-2025-1234",
  "source": "nvd",
  "product": "apache",
  "version": "2.4.49",
  "description": "RCE in Apache...",
  "cvss": 9.8,
  "published": "2025-06-28",
  "references": ["https://..."],
  "exploit_available": true,
  "cpe_guess": "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*"
}
```

### 2. Nguồn hỗ trợ hiện tại:

* OpenCVE / NVD JSON
* Threatpost RSS
* GitHub issue RSS
* HackerNews

### 3. Tiêu chí chuẩn hóa:

| Trường               | Diễn giải                                              |
| -------------------- | ------------------------------------------------------ |
| `id`                 | CVE hoặc custom ID (ZERO-xxxx)                         |
| `product`, `version` | Phần mềm bị ảnh hưởng                                  |
| `cvss`               | Nếu không có → gán theo rule mapping                   |
| `references`         | Các link kỹ thuật, PoC, blog                           |
| `exploit_available`  | Đúng nếu có PoC, khai thác                             |
| `cpe_guess`          | Có thể tính từ product/version nếu không có chính thức |

### 4. Hàm normalize đề xuất:

```go
func NormalizeFromThreatPost(rssItem RSSItem) UVF {
    return UVF{
        ID: GenerateCustomID(rssItem),
        Product: ExtractProduct(rssItem.Title),
        Version: ExtractVersion(rssItem.Summary),
        Description: rssItem.Summary,
        Source: "threatpost",
        References: []string{rssItem.Link},
    }
}
```

Hướng dẫn viết cho từng source nên để trong `/pkg/feeds/<source>.go`

---

Tài liệu `architecture.md`, `feed_formats.md` và `vmp.md` giúp team onboard nhanh, dễ test unit và phát triển module parser, dashboard theo chuẩn chung.
