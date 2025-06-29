# vulnsense
VulnSense – Your early-warning system for vulnerabilities.

> 🔐 A modular, Go-based threat intelligence engine for detecting and matching vulnerabilities against real assets.

## 🧠 Overview

VulnSense is a vulnerability intelligence and alerting system that:
- Collects vulnerability data from multiple sources (CVE, blogs, RSS, GitHub)
- Normalizes into a unified vulnerability format (UVF)
- Matches vulnerabilities with assets from Splunk inventory
- Sends alerts via Slack, Email, or SIEM
- Supports future extensions: risk triage, patch tracking, ASM modules

---

## 🚀 Quick Start

```bash
git clone https://github.com/your-org/vulnsense.git
cd vulnsense
go build -o vulnsense cmd/main.go
./vulnsense --config configs/app.yaml
```

## 📁 Project Structure

```bash
/vulnsense
├── cmd/                # Entrypoint (main.go)
├── pkg/
│   ├── feeds/          # Feed fetchers (OpenCVE, RSS, GitHub)
│   ├── parser/         # Normalize to UVF
│   ├── matcher/        # Match assets vs vulnerabilities
│   ├── alert/          # Slack, email alerts
│   ├── splunk/         # Splunk asset querying
│   └── store/          # Database/cache access
├── internal/           # Runner orchestration
├── configs/            # YAML configuration files
├── scripts/            # Data seeds / setup
├── web/                # Optional web UI dashboard
└── README.md
```

## 📝 Configuration

Example configs/app.yaml:

```bash
splunk:
  url: "https://splunk.internal"
  token: "your-token"
feeds:
  schedule: "30m"
alert:
  slack_webhook: "https://hooks.slack.com/..."
```

## 🧪 Run Tests

```bash
go test ./pkg/... -v
```

## 📚 Documentation
- docs/architecture.md - System design

- docs/feed_formats.md - Unified vulnerability format (UVF)

- docs/vmp.md - Vulnerability management features (triage, SLA, patch tracking)

## 📄 License
This project is licensed under the MIT License. See LICENSE file for details.

## ✨ Contributors
Project lead: 0xmanhnv@gmail.com

Docs & specs by: VulnSense team

