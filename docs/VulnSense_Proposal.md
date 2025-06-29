# ĐỀ XUẤT DỰ ÁN: HỆ THỐNG PHÁT HIỆN VÀ CẢNH BÁO LỖ HỖNG BẢO MẬt TỬ ĐỘNG (VULNSENSE)

## 1. Mục Đích

Trong bối cảnh lỗ hổng bảo mật gia tăng nhanh chóng, nhiều lỗ hổng đã bị khai thác ngay khi công bố và chưa có CVE, việc theo dõi thông tin lỗ hổng và đối chiếu với hệ thống của tổ chức là yêu cầu bắt buộc. Tuy nhiên, các công cụ hiện tại thường chỉ giới hạn với CVE đã được chuẩn hóa.

**VulnSense** được đề xuất như một nền tảng thu thập thông tin lỗ hổng từ nhiều nguồn khác nhau (có CVE & không CVE), chuẩn hóa, so khớp với danh sách tài sản thực tế và cảnh báo nguy cơ theo thời gian thật.

---

## 2. Mục Tiêu

* Xây dựng hệ thống gom thông tin lỗ hổng bảo mật từ nhiều nguồn:

  * NVD / OpenCVE
  * Threat Intelligence / Blog / GitHub / Zero-day
* Chuẩn hóa thông tin lỗ hổng về định dạng chuẩn.
* Tự động so khớp với asset inventory từ Splunk.
* Phát hiện nhanh các thiết bị/phần mềm bị ảnh hưởng.
* Cảnh báo qua Slack, email, SIEM.

---

## 3. Lợi Ích Mang Lại

| Lợi ích            | Mô tả                                                        |
| ------------------ | ------------------------------------------------------------ |
| 🚨 Phát hiện sớm   | Ngay khi lỗ hổng được công khai trên blog/forum, chưa có CVE |
| 🌎 Tổng hợp nguồn  | Kế hợp các nguồn CVE và phi CVE (đa chiều)                   |
| 📊 So khớp thực tế | Tự động liên hệ asset thực tế trong Splunk                   |
| 📢 Cảnh báo nhanh  | Qua Slack/email khi match nguy cơ                            |
| 🚀 Có định hướng   | Mở rộng lên ASM (Attack Surface Management)                  |

---

## 4. Tổng Quan Kiến Trúc

* **Feed Engine**: gom thông tin từ OpenCVE, threat blog, RSS, GitHub
* **Parser/Normalizer**: convert về dạng chuẩn
* **VulnStore**: DB lỗ hổng, truy vấn nhanh
* **Matcher Engine**: so khớp với asset Splunk
* **Alert Engine**: Slack, Email, Webhook

> Hệ thống viết bằng Go (modular), deploy Docker/K8s, chuẩn CI/CD

---

## 5. Giai Đoạn Triển Khai

| Giai đoạn | Mô tả                                     | Thời gian   |
| --------- | ----------------------------------------- | ----------- |
| P1        | Core CVE + ZeroDay Feed + Splunk match    | 1 tháng     |
| P2        | Alert Slack + Email + UI Dashboard cơ bản | 1-2 tuần    |
| P3        | Chuẩn hóa feed, log, CI/CD                | 2 tuần      |
| P4        | Mở rộng ASM (Recon, Subdomain, Portscan)  | sau 3 tháng |

---

## 6. Nhân Sự & Tài Nguyên

* 01 Leader (Tech/Product)
* 01 Dev backend (Go)
* 01 SecOps support (Splunk API, alert logic)
* Infra: 1 server Docker / Kubernetes

---

## 7. Tên Dự Án Đề Xuất: **VulnSense**

> "A unified vulnerability sensing & alerting platform."

---

## 8. Kiến Nghi

* Phê duyệt triển khai giai đoạn POC trong 4 tuần
* Cho phép truy cập API Splunk + repo OpenCVE
* Xem xét mở rộng roadmap lên ASM full trong 6 tháng

---

*Tài liệu này dùng để trình bày với lãnh đạo CNTT, An ninh và Quản trị ban dự án. Có thể đính kèm demo UI hoặc slide khi bàn giao.*
