# Enterprise News Portal (Portal Berita Modern)

> Portal berita modern berkecepatan tinggi, dinamis, dan scalable (setara CNN / Detik) dengan arsitektur **Go 1.23+**, **MySQL 8 / MariaDB 11**, **Redis 7**, **Astro SSR** (Frontend Publik), dan **React + shadcn/ui** (Panel Admin).

---

## 🚀 Prinsip Utama: 100% Konten Dinamis

- **Astro SSR Penuh (`output: 'server'`)**: Tidak ada Static Site Generation (SSG) untuk berita. Semua konten diambil saat request dari API Go via jaringan internal.
- **Tanpa Rebuild & Tanpa Deploy**: Artikel baru, edit, dan unpublish tampil di homepage, kategori, sitemap, dan RSS dalam **≤ 5 detik**.
- **Multi-tier Caching**: L1 Cloudflare CDN, L2 Redis dengan `singleflight` (anti cache stampede), L3 in-process memory.
- **Realtime Trending & View Counter**: Deduplikasi kunjungan per IP+UA, Redis atomic counter & ZSET windowing, flush berkala ke MySQL.

---

## 📁 Struktur Monorepo

```
news-portal/
├── apps/
│   ├── api/          # Go 1.23+ Clean Architecture Backend
│   ├── web/          # Astro SSR Frontend Publik (Template Tailnews)
│   └── admin/        # React 19 + Vite + TanStack Router (shadcn-admin)
├── deploy/           # Docker Compose, Nginx Reverse Proxy, .env examples
├── tests/            # E2E Playwright & Load Test k6
├── docs/             # architecture.md, openapi.yaml, caching.md, runbook.md
├── Makefile          # Automation commands
└── README.md
```

---

## 🛠️ Prasyarat (Prerequisites)

- **Go 1.23+**
- **Node.js v20+**
- **MySQL 8** atau **MariaDB 11.8+** (port 3306)
- **Redis 7** (port 6379)
- **Docker & Docker Compose**

---

## ⚡ Memulai Cepat (Quick Start)

### 1. Menjalankan Lokal

```bash
# 1. Pastikan dependensi Go & env terkonfigurasi
cp apps/api/.env.example apps/api/.env

# 2. Jalankan seluruh layanan via Makefile
make dev
```

Endpoint yang tersedia:
- **API Health Check**: `http://localhost:8080/healthz`
- **API Readiness Check**: `http://localhost:8080/readyz`
- **API Namespace**: `http://localhost:8080/api/v1/ping`

### 2. Pengujian & Linting

```bash
# Menjalankan unit & integration tests backend
make test

# Menjalankan linter & static analysis
make lint

# Kompilasi binary produksi
make build
```

---

## 📖 Dokumentasi Lengkap

- [Spesifikasi Arsitektur](docs/architecture.md)
- [OpenAPI 3.0 Specification](docs/openapi.yaml) *(Fase 2)*
- [Strategi Caching & Cloudflare Rules](docs/caching.md) *(Fase 3)*
- [Baseline Performa & k6 Benchmark](docs/performance.md) *(Fase 3)*
- [Runbook Operasional](docs/runbook.md) *(Fase 6)*
