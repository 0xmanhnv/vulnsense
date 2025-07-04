# RSS Feeds Integration

## Tổng quan

VulnSense hiện tại đã hỗ trợ đọc và xử lý RSS feeds từ các nguồn security blogs, advisory feeds, và vulnerability databases. Hệ thống sử dụng thư viện **gofeed** để parse RSS/Atom feeds và tự động:

1. Đọc danh sách RSS feeds từ `configs/sources.yaml`
2. Fetch nội dung từ các RSS feeds với gofeed library
3. Parse và normalize thành format vulnerability chuẩn
4. Convert và lưu vào database để matching với assets

## Cấu hình RSS Feeds

Thêm các RSS feeds vào file `configs/sources.yaml`:

```yaml
- name: "BleepingComputer"
  url: https://www.bleepingcomputer.com/feed/
  type: news
  source: BleepingComputer
  enabled: true

- name: "KrebsOnSecurity"
  url: https://krebsonsecurity.com/feed
  type: blog
  source: KrebsOnSecurity
  enabled: true

- name: "CISA Advisories"
  url: https://www.cisa.gov/cybersecurity-advisories/all.xml
  type: advisory
  source: CISA
  enabled: true

- name: "VulDB Recent"
  url: https://vuldb.com/?rss.recent
  type: vulnerability
  source: VulDB
  enabled: true
```

### Các trường bắt buộc:

- **`name`**: Tên hiển thị của feed
- **`url`**: URL của RSS feed  
- **`type`**: Loại feed (xem bên dưới)
- **`source`**: Identifier của nguồn
- **`enabled`**: Enable/disable feed (true/false)

### Các loại feed types:

- **`blog`**: Security blogs (KrebsOnSecurity, Graham Cluley)
- **`news`**: Security news sites (BleepingComputer, SecurityWeek)
- **`advisory`**: Official advisories (CISA, vendor advisories)
- **`vulnerability`**: Vulnerability databases (VulDB, CVEFeed.io)
- **`exploit`**: Exploit databases (ExploitDB, PacketStorm)
- **`research`**: Security research (Google Project Zero, Trail of Bits)
- **`vendor-advisory`**: Vendor-specific advisories (Microsoft MSRC)
- **`cve`**: Dedicated CVE feeds

## Architecture

### Clean Architecture Implementation

```
configs/sources.yaml
       ↓
internal/config/config.go (RSSSource struct)
       ↓  
internal/adapter/feeds.go (NewFeedProviders)
       ↓
internal/adapter/rss_feed.go (RSSFeedAdapter + gofeed)
       ↓
pkg/feed/fetcher.go (Vulnerability type - không depend internal)
       ↓
internal/adapter/feed_converter.go (convert pkg → domain)
       ↓
internal/usecase/fetch_vulnerabilities.go
```

### Dependency Flow

- **`pkg/feed/`**: Public interfaces, NO dependency on internal packages
- **`internal/adapter/`**: Implementation using external libraries (gofeed)
- **Converter Layer**: Bridges pkg types ↔ domain types
- **Domain Layer**: Core business logic

## Features

### 🎯 **Smart Content Filtering**
- Tự động lọc nội dung security-related
- 20+ security keywords detection
- Loại bỏ spam và noise từ feeds
- Feed type prioritization (vulnerability/advisory feeds always included)

### 🔍 **CVE Detection & Normalization**
- Regex pattern: `(?i)CVE-\d{4}-\d{4,}`
- Tự động extract và uppercase CVE IDs
- Preferential ID generation (CVE > GUID > Link hash)

### 🏷️ **Product Extraction** 
- 15+ platform detection (Apache, Nginx, MySQL, PostgreSQL, Redis, Docker, etc.)
- Regex pattern matching: `(?i)\b(apache|nginx|mysql|...)\b`
- Smart fallback to "Unknown"

### ⚡ **Intelligent Severity Scoring**
- **CRITICAL**: "critical", "rce", "remote code execution"
- **HIGH**: "high", "privilege escalation", "authentication bypass"  
- **MEDIUM**: "medium", "xss", "sql injection"
- **LOW**: "low", "information disclosure"
- Feed type consideration (exploit feeds default HIGH)

### 🔗 **Exploit Detection**
- Keywords: "exploit", "poc", "proof of concept", "metasploit", "exploit-db"
- Active exploitation: "in the wild", "weaponized", "publicly available"
- Boolean flag cho exploit availability

### 🏷️ **Auto Tag Generation**
- Technology detection: web, database, network, crypto, auth
- Platform detection: windows, linux, macos, mobile, cloud, container
- Feed type as tag

## Cách chạy

### Test RSS feeds:

```bash
go run cmd/test-rss/main.go
```

Output mẫu:
```
=== Testing BleepingComputer ===
Found 12 potential vulnerabilities from BleepingComputer
  1. CVE-2024-12345 [HIGH] - Critical Windows vulnerability
     Product: windows | CVEs: [CVE-2024-12345]
     Tags: [windows news]
```

### Production:

```bash
# Chạy scheduler cho auto-fetch
go run cmd/scheduler/main.go

# Chạy main application  
go run cmd/vulnsense/main.go
```

## Dependencies

- **github.com/mmcdole/gofeed v1.3.0**: Universal feed parser (RSS/Atom/JSON)
- **net/http**: HTTP client với 30s timeout
- **regexp**: CVE và product pattern matching
- **time**: Date parsing và formatting

### Gofeed Library Benefits:

- **Universal**: RSS 0.90-2.0, Atom 0.3-1.0, JSON 1.0-1.1
- **Robust**: Handles broken/invalid XML feeds
- **Extensions**: Support for Dublin Core, iTunes extensions
- **Performance**: Production-ready với comprehensive error handling

## Monitoring & Logs

RSS fetching sẽ log:

```bash
INFO Starting RSS feed fetch url=https://krebsonsecurity.com/feed feed_name=KrebsOnSecurity
INFO Successfully fetched RSS vulnerabilities count=15 source=KrebsOnSecurity
INFO Initialized feed providers count=25 rss_feeds=20
```

## Troubleshooting

### Common Issues:

1. **RSS feed timeout (30s)**
   ```bash
   # Adjust timeout in RSSFeedAdapter constructor
   Timeout: 60 * time.Second
   ```

2. **Invalid XML/parsing errors**
   - gofeed handles most broken feeds automatically
   - Check feed URL manually: `curl -s "URL" | head -50`

3. **No vulnerabilities detected**
   - Check `isSecurityRelated()` keywords
   - Verify feed type configuration  
   - Enable more feeds with `enabled: true`

4. **Too many false positives**
   - Adjust security keywords in `securityKeywords` slice
   - Tune feed type logic

### Debug Commands:

```bash
# Test specific feed
go run cmd/test-rss/main.go

# Check dependencies
go mod tidy && go mod verify

# Validate YAML config
cat configs/sources.yaml | python3 -c "import yaml,sys; yaml.safe_load(sys.stdin)"
```

## Extending RSS Support

### Thêm feed type mới:
1. Update `determineSeverity()` function in `rss_feed.go`
2. Adjust `isSecurityRelated()` logic nếu cần
3. Thêm vào `sources.yaml`

### Thêm product detection:
```go
// Update prodPattern in NewRSSFeedAdapter
prodPattern: regexp.MustCompile(`(?i)\b(apache|nginx|mysql|...|newproduct)\b`),
```

### Custom feed parsing:
1. Tạo custom adapter implement `feed.Fetcher`
2. Use converter pattern cho domain integration
3. Register trong `NewFeedProviders()`

## Performance

- **Concurrent**: Mỗi RSS feed processed tuần tự hiện tại
- **Memory**: ~1-5MB per feed depending on item count
- **HTTP timeout**: 30s per feed request  
- **Parsing**: gofeed library optimized cho production
- **Database**: Bulk insert thông qua usecase layer

## Best Practices

1. **Feed Selection**: Chọn feeds có quality content, avoid noise
2. **Enable/Disable**: Sử dụng `enabled` flag để test feeds
3. **Monitoring**: Check logs cho parsing errors  
4. **Testing**: Luôn test new feeds với `cmd/test-rss`
5. **Rate Limiting**: Consider RSS provider rate limits 