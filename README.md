# TailNews

TailNews adalah monorepo portal berita yang berisi API, situs publik, dan panel administrasi. Situs publik menggunakan Astro SSR, API menggunakan Go, dan panel administrasi dibangun dengan React.

## Komponen

| Aplikasi | Teknologi | Direktori |
| --- | --- | --- |
| API | Go, MySQL/MariaDB, Redis | `apps/api` |
| Situs publik | Astro SSR | `apps/web` |
| Panel administrasi | React, Vite, TanStack Router | `apps/admin` |

Konfigurasi deployment Docker Compose dan Nginx tersedia di `deploy/`. Dokumentasi arsitektur dan spesifikasi API berada di `docs/`.

## Persyaratan

- Docker dan Docker Compose untuk menjalankan seluruh stack.
- Untuk pengembangan tanpa container aplikasi: Go sesuai versi pada `apps/api/go.mod`, serta Node.js 20 atau lebih baru.
- MySQL atau MariaDB dan Redis jika menjalankan API secara lokal.

## Menjalankan dengan Docker Compose

Perintah berikut membangun dan menjalankan API, situs publik, panel admin, MariaDB, Redis, dan Nginx:

```bash
make docker-up
```

Alamat layanan:

- Situs publik: <http://localhost:4321>
- Panel administrasi: <http://localhost:5173>
- API health check: <http://localhost:8080/healthz>
- Nginx: <http://localhost>

Skema database perlu dimigrasikan sebelum menggunakan fitur yang bergantung pada data. Migrasi menggunakan Goose dan target Makefile berikut:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
make migrate-up
```

Untuk menghentikan stack:

```bash
make docker-down
```

## Pengembangan Lokal

Jalankan MariaDB dan Redis terlebih dahulu. Salin konfigurasi contoh, lalu sesuaikan DSN database dan secret JWT pada `apps/api/.env`:

```bash
cp apps/api/.env.example apps/api/.env
```

Setelah database tersedia dan Goose terpasang, terapkan migrasi:

```bash
make migrate-up
```

Jalankan setiap perintah berikut di terminal terpisah:

```bash
make dev-api
make dev-web
make dev-admin
```

API, situs publik, dan panel admin menggunakan port `8080`, `4321`, dan `5173` secara default.

## Pengujian dan Build

Perintah berikut menjalankan pengujian backend, pemeriksaan statis Go, dan membangun binary API:

```bash
make test
make lint
make build
```

Build situs publik dan panel admin dapat dijalankan dari direktori masing-masing dengan `npm run build`.

## Struktur Repository

```text
apps/
	admin/       Panel administrasi
	api/         Backend dan migrasi database
	web/         Situs publik
deploy/        Docker Compose, Nginx, dan contoh konfigurasi
docs/          Dokumentasi proyek
tests/         Pengujian dan benchmark
```

## Dokumentasi

- [Arsitektur](docs/architecture.md)
- [Spesifikasi API](docs/openapi.yaml)
- [Aturan Cloudflare](docs/cloudflare-rules.md)
