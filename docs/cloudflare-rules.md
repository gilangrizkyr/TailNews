# Cloudflare Edge CDN & Cache Architecture

Dokumentasi spesifikasi konfigurasi Cloudflare CDN dan Edge Rules untuk portal berita **NewsPortal**, dirancang untuk memenuhi target:
- **TTFB < 100ms** dari edge cache di seluruh Indonesia (CDN PoP Jakarta/Surabaya/Makassar).
- $\ge$ **95% request pembaca** disajikan langsung dari Edge Cache tanpa menyentuh server asal (Origin Shielding).
- **Zero SSG**: Konten tetap 100% dinamis melalui Astro SSR + Go API, dengan propagasi pembaruan berita $\le$ 5 detik melalui Cache Tagging & Purge API.

---

## 1. Cloudflare Page Rules & Cache Rules (Terraform / Dashboard)

### Rule 1: Static Media Assets & Uploads
- **Expression**:
  ```text
  (http.request.uri.path rmatches "^/media/" or http.request.uri.path rmatches "\.(webp|png|jpe?g|svg|ico|woff2?|css|js)$")
  ```
- **Settings**:
  - **Eligible for cache**: Yes
  - **Edge Cache TTL**: 30 days (2,592,000s)
  - **Browser Cache TTL**: 30 days (`public, max-age=2592000, immutable`)
  - **Origin Cache Control**: Respect headers
  - **Serve Stale Content**: While revalidating (stale-while-revalidate = 86,400s)

---

### Rule 2: Bypass Cache for Admin, Auth & Preview Tokens
- **Expression**:
  ```text
  (http.request.uri.path rmatches "^/(admin|api/v1/admin|api/v1/auth)/" or
   http.request.uri.query contains "preview=" or
   http.cookie contains "newsroom_auth" or
   http.request.method in {"POST" "PUT" "DELETE" "PATCH"})
  ```
- **Settings**:
  - **Bypass Cache**: Yes
  - **Origin Cache Control**: Disabled
  - **Security Level**: High
  - **WAF**: Enabled

---

### Rule 3: Dynamic Reader Pages (Homepage, Category, Author, Tag, Detail Berita)
- **Expression**:
  ```text
  (http.request.uri.path eq "/" or
   http.request.uri.path rmatches "^/(category|tag|author|page)/" or
   http.request.uri.path rmatches "^/[a-z0-9-]+$" or
   http.request.uri.path in {"/sitemap.xml" "/news-sitemap.xml" "/rss.xml"}) and
   not (http.request.uri.query contains "preview=")
  ```
- **Settings**:
  - **Eligible for cache**: Yes
  - **Edge Cache TTL**: 120 seconds
  - **Browser Cache TTL**: 10 seconds
  - **Stale While Revalidate**: 60 seconds
  - **Stale If Error**: 300 seconds (Origin Protection saat server maintenance)
  - **Custom Cache Key**:
    - Include: Scheme, Host, Path, Query string
    - Ignore: Tracking params (`utm_*`, `fbclid`, `gclid`)

---

## 2. Cloudflare Instant Purge API Integration

Ketika editor mempublikasikan atau mengubah berita di panel admin, Go backend API memicu pembersihan cache via Cloudflare API v4 secara asynchronous:

```bash
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CF_ZONE_ID}/purge_cache" \
     -H "Authorization: Bearer ${CF_API_TOKEN}" \
     -H "Content-Type: application/json" \
     -d '{
       "files": [
         "https://newsportal.id/",
         "https://newsportal.id/rss.xml",
         "https://newsportal.id/news-sitemap.xml",
         "https://newsportal.id/'"${ARTICLE_SLUG}"'",
         "https://newsportal.id/category/'"${CATEGORY_SLUG}"'"
       ]
     }'
```

---

## 3. Web Application Firewall (WAF) & Rate Limiting

### WAF Rule 1: Login Brute Force Protection
- **Target**: `POST /api/v1/auth/login`
- **Threshold**: > 10 requests dalam 10 detik per IP.
- **Action**: Block IP selama 15 menit.

### WAF Rule 2: API Abuse Protection
- **Target**: `GET /api/v1/*`
- **Threshold**: > 300 requests dalam 1 menit per IP.
- **Action**: Managed Challenge (Turnstile Captcha).

### WAF Rule 3: Anti-Scraping & Bad Bots
- **Settings**:
  - **Bot Fight Mode**: ON
  - **Block AI Scrapers**: ON (melindungi konten jurnalis dari scraping liar)
  - **HTTP/3 (QUIC)**: Enabled
  - **0-RTT Connection Resumption**: Enabled
  - **Early Hints (103)**: Enabled (untuk preloading font Inter dan CSS kritis)

---

## 4. Arsitektur Origin Shield & DNS

```
[ Reader Request ]
       │
       ▼ (Anycast DNS & TLS 1.3 Termination)
[ Cloudflare Edge PoP (Jakarta/Surabaya) ]
  ├── Cache Hit (95%+ traffic, TTFB < 50ms) ──────────┐
  └── Cache Miss / Dynamic SSR                         │
       │                                               │
       ▼ (Authenticated Origin Pulls)                  │
[ Nginx Reverse Proxy (Port 80/443) ]                  │
  ├── /media/*  ──> Local Storage / Cloudflare R2      │
  ├── /api/*    ──> Go Backend API (Port 8080)         │
  └── /*        ──> Astro Node SSR (Port 4321)         │
                                                       │
                                 ◄─────────────────────┘
```
