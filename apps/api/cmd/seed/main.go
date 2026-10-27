package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/argon2"
	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/platform"
)

func hashPassword(password string) string {
	salt := []byte("salt-for-dev-seeding-1234567890")
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%x$%x", salt, hash)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	ctx := context.Background()
	db, err := platform.NewMySQLPool(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("==> Starting database seeding...")
	start := time.Now()

	// 1. Seed Roles & Users
	defaultPassword := hashPassword("AdminSecret123!")

	// Insert Users
	_, _ = db.Exec(`INSERT IGNORE INTO users (id, role_id, name, email, password_hash, is_active) VALUES
		(1, 1, 'Super Admin', 'admin@newsportal.id', ?, 1),
		(2, 2, 'Editor Utama', 'editor@newsportal.id', ?, 1),
		(3, 2, 'Budi Santoso', 'budi.editor@newsportal.id', ?, 1),
		(4, 3, 'Rina Wijaya', 'rina.writer@newsportal.id', ?, 1),
		(5, 3, 'Doni Pratama', 'doni.writer@newsportal.id', ?, 1),
		(6, 3, 'Siti Rahma', 'siti.writer@newsportal.id', ?, 1);`,
		defaultPassword, defaultPassword, defaultPassword, defaultPassword, defaultPassword, defaultPassword,
	)

	// 2. Seed Categories
	categories := []struct {
		id        int
		name      string
		slug      string
		sortOrder int
	}{
		{1, "Politik", "politik", 1},
		{2, "Ekonomi", "ekonomi", 2},
		{3, "Teknologi", "teknologi", 3},
		{4, "Olahraga", "olahraga", 4},
		{5, "Hiburan", "hiburan", 5},
		{6, "Otomotif", "otomotif", 6},
	}
	for _, c := range categories {
		_, err := db.Exec(`INSERT INTO categories (id, name, slug, description, sort_order, is_nav)
			VALUES (?, ?, ?, ?, ?, 1)
			ON DUPLICATE KEY UPDATE name=VALUES(name), sort_order=VALUES(sort_order);`,
			c.id, c.name, c.slug, fmt.Sprintf("Berita terkini seputar %s nasional dan internasional.", c.name), c.sortOrder,
		)
		if err != nil {
			log.Fatalf("failed to seed category %s: %v", c.name, err)
		}
	}

	// 3. Seed Media
	mediaItems := []struct {
		id       int
		filename string
		path     string
		alt      string
	}{
		{1, "hero_politics.webp", "/media/hero_politics.webp", "Presiden berpidato di hadapan parlemen"},
		{2, "stock_market.webp", "/media/stock_market.webp", "Layar pergerakan Indeks Harga Saham Gabungan"},
		{3, "ai_datacenter.webp", "/media/ai_datacenter.webp", "Server komputasi kecerdasan buatan"},
		{4, "football_match.webp", "/media/football_match.webp", "Pertandingan seru sepak bola di stadion utama"},
		{5, "concert_stage.webp", "/media/concert_stage.webp", "Panggung konser musik spektakuler"},
		{6, "electric_car.webp", "/media/electric_car.webp", "Peluncuran mobil listrik konsep masa depan"},
	}
	for _, m := range mediaItems {
		variants := fmt.Sprintf(`{"thumbnail":"%s","medium":"%s","large":"%s"}`, m.path, m.path, m.path)
		_, err := db.Exec(`INSERT INTO media (id, filename, file_path, mime_type, size_bytes, width, height, alt_text, caption, photographer, variants)
			VALUES (?, ?, ?, 'image/webp', 245000, 1200, 800, ?, ?, 'Tim Dokumentasi Redaksi', ?)
			ON DUPLICATE KEY UPDATE alt_text=VALUES(alt_text);`,
			m.id, m.filename, m.path, m.alt, m.alt, variants,
		)
		if err != nil {
			log.Fatalf("failed to seed media: %v", err)
		}
	}

	// 4. Seed Authors
	authors := []struct {
		id     int
		userID int
		name   string
		slug   string
		bio    string
	}{
		{1, 4, "Rina Wijaya", "rina-wijaya", "Jurnalis investigasi dengan pengalaman 10 tahun meliput isu politik dan kebijakan publik."},
		{2, 5, "Doni Pratama", "doni-pratama", "Analis ekonomi dan pasar modal, rutin mengulas pergerakan makroekonomi."},
		{3, 6, "Siti Rahma", "siti-rahma", "Pemerhati tren teknologi kecerdasan buatan, gadget, dan industri digital."},
		{4, 2, "Budi Santoso", "budi-santoso", "Editor olahraga dan kolumnis sepak bola nasional dan liga internasional."},
		{5, 3, "Ari Budin", "ari-budin", "Fotografer dan jurnalis feature budaya, gaya hidup, dan otomotif."},
	}
	for _, a := range authors {
		avatarID := 1
		_, err := db.Exec(`INSERT INTO authors (id, user_id, name, slug, bio, avatar_media_id, social_links)
			VALUES (?, ?, ?, ?, ?, ?, '{"twitter":"https://twitter.com","linkedin":"https://linkedin.com"}')
			ON DUPLICATE KEY UPDATE name=VALUES(name), bio=VALUES(bio);`,
			a.id, a.userID, a.name, a.slug, a.bio, avatarID,
		)
		if err != nil {
			log.Fatalf("failed to seed author: %v", err)
		}
	}

	// 5. Seed Tags
	tags := []string{
		"Pemilu", "Kecerdasan Buatan", "Startup", "Gadget", "Timnas Indonesia",
		"Bursa Efek", "Mobil Listrik", "Investasi", "Film Indonesia", "Wisata",
		"Ekonomi Digital", "Piala Dunia", "Kemenkeu", "Inovasi", "Kesehatan",
	}
	tagIDs := make([]int, len(tags))
	for i, t := range tags {
		slug := strings.ToLower(strings.ReplaceAll(t, " ", "-"))
		res, err := db.Exec(`INSERT INTO tags (name, slug) VALUES (?, ?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id);`, t, slug)
		if err != nil {
			log.Fatalf("failed to seed tag: %v", err)
		}
		id, _ := res.LastInsertId()
		tagIDs[i] = int(id)
	}

	// 6. Seed Static Pages
	pages := []struct {
		title string
		slug  string
		body  string
	}{
		{"Tentang Kami", "tentang-kami", "<h2>Tentang Portal Berita</h2><p>Kami adalah portal media siber terpercaya yang menyajikan berita aktual, cepat, dan berimbang secara profesional sesuai Kode Etik Jurnalistik.</p>"},
		{"Pedoman Media Siber", "pedoman-media-siber", "<h2>Pedoman Pemberitaan Media Siber</h2><p>Kemerdekaan berpendapat, kemerdekaan berekspresi, dan kemerdekaan pers adalah hak asasi manusia yang dilindungi Pancasila, Undang-Undang Dasar 1945, dan Piagam Perserikatan Bangsa-Bangsa tentang Hak Asasi Manusia.</p>"},
		{"Susunan Redaksi", "redaksi", "<h2>Susunan Redaksi & Dewan Redaksi</h2><p>Pemimpin Redaksi: Editor Utama<br/>Redaktur Pelaksana: Budi Santoso<br/>Dewan Penasihat: Dewan Pakar Media Nasional.</p>"},
		{"Kebijakan Privasi", "kebijakan-privasi", "<h2>Kebijakan Privasi</h2><p>Kebijakan Privasi ini menjelaskan bagaimana kami mengumpulkan, menggunakan, dan melindungi data pribadi pembaca saat mengakses portal berita kami.</p>"},
	}
	for _, p := range pages {
		_, err := db.Exec(`INSERT INTO pages (title, slug, content, is_published)
			VALUES (?, ?, ?, 1)
			ON DUPLICATE KEY UPDATE content=VALUES(content);`,
			p.title, p.slug, p.body,
		)
		if err != nil {
			log.Fatalf("failed to seed page: %v", err)
		}
	}

	// 7. Seed >= 1,000 Articles in Batches
	var existingArticlesCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&existingArticlesCount)
	if existingArticlesCount >= 1000 {
		log.Printf("==> Database already seeded with %d articles. Skipping duplicate insert to preserve data integrity.\n", existingArticlesCount)
		return
	}

	log.Println("==> Seeding >= 1,000 realistic articles in batches...")

	titleTemplates := []string{
		"Pemerintah Resmi Luncurkan Kebijakan Stimulus Baru untuk Sektor %s",
		"Inovasi Terbaru di Bidang %s Mampu Tingkatkan Produktivitas Nasional",
		"Perkembangan Pesat Sektor %s Menarik Minat Investor Global",
		"Sorotan Utama: Analisis Strategis Masa Depan Industri %s di Asia",
		"Tantangan dan Peluang Baru Menghadapi Dinamika %s Tahun Ini",
		"Langkah Nyata Transformasi Modern pada Sektor %s Berkelanjutan",
		"Pakar Menilai Penguatan Regulasi %s Krusial bagi Pertumbuhan Ekonomi",
		"Kesiapan Infrastruktur Mendukung Akselerasi Pertumbuhan %s di Indonesia",
		"Kisah Inspiratif Para Pelaku Usaha yang Sukses Menggebrak Bidang %s",
		"Kajian Mendalam: Mengapa Tren %s Menjadi Sorotan Publik Pekan Ini",
	}

	sampleBodies := []string{
		"<p>Perkembangan teknologi dan dinamika pasar saat ini mendorong berbagai pemangku kepentingan untuk merumuskan strategi adaptif yang berdaya saing tinggi. Dalam pertemuan tahunan yang diselenggarakan di Jakarta, para ahli menegaskan pentingnya kolaborasi lintas sektor guna memastikan keberlanjutan ekosistem industri.</p><h2>Langkah Strategis ke Depan</h2><p>Tidak hanya berbicara mengenai target jangka pendek, fokus utama kini diarahkan pada penguatan fundamental tata kelola, transparansi, serta integrasi teknologi modern. Hal ini terbukti mampu mengoptimalkan efisiensi operasional hingga 35 persen dibandingkan metode konvensional.</p><blockquote>Inovasi yang berkelanjutan adalah kunci utama menjaga ketahanan nasional di tengah ketidakpastian ekonomi global.</blockquote><p>Masyarakat diharapkan terus proaktif dalam mengikuti perkembangan informasi terverifikasi agar dapat memanfaatkan setiap peluang dengan optimal.</p>",
		"<p>Dalam era digital yang serba cepat, kecepatan dan akurasi informasi menjadi pilar penentu keberhasilan implementasi program strategis. Evaluasi triwulan terbaru menunjukkan indikator positif yang mencerminkan kepercayaan publik yang kian solid terhadap pemulihan ekonomi nasional.</p><h2>Peningkatan Kapasitas dan Daya Saing</h2><p>Sejumlah terobosan baru diperkenalkan guna mempercepat adopsi teknologi ramah lingkungan dan inklusif. Menurut data yang dirilis pekan lalu, terjadi peningkatan partisipasi aktif dari generasi muda sebesar 42 persen dalam sektor-sektor produktif.</p><p>Ke depan, penguatan sinergi antara pemerintah, akademisi, dan pelaku usaha akan terus diperkuat guna menciptakan ekosistem yang sehat dan berkeadilan.</p>",
	}

	totalToSeed := 1020
	batchSize := 100
	r := rand.New(rand.NewSource(42))

	now := time.Now().UTC()
	var articleIDs []int64

	for batchStart := 0; batchStart < totalToSeed; batchStart += batchSize {
		count := batchSize
		if batchStart+count > totalToSeed {
			count = totalToSeed - batchStart
		}

		valueStrings := make([]string, 0, count)
		valueArgs := make([]interface{}, 0, count*15)

		for i := 0; i < count; i++ {
			articleNum := batchStart + i + 1
			cat := categories[r.Intn(len(categories))]
			author := authors[r.Intn(len(authors))]
			mediaID := mediaItems[r.Intn(len(mediaItems))].id

			template := titleTemplates[r.Intn(len(titleTemplates))]
			title := fmt.Sprintf(template, cat.name)
			if articleNum > 10 {
				title = fmt.Sprintf("%s (Bagian %d)", title, articleNum)
			}
			slug := fmt.Sprintf("%s-%d", strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(title, " ", "-"), ":", "")), articleNum)
			excerpt := fmt.Sprintf("Ringkasan komprehensif mengenai %s dan dampaknya terhadap perkembangan terkini di Indonesia.", title)
			body := sampleBodies[r.Intn(len(sampleBodies))]

			// Determine status
			status := "published"
			if articleNum%45 == 0 {
				status = "draft"
			} else if articleNum%50 == 0 {
				status = "review"
			} else if articleNum%55 == 0 {
				status = "scheduled"
			}

			// PublishedAt: spread across last 60 days
			minutesAgo := r.Intn(60 * 24 * 60)
			pubAt := now.Add(-time.Duration(minutesAgo) * time.Minute)
			if status == "scheduled" {
				pubAt = now.Add(time.Duration(r.Intn(120)+10) * time.Minute)
			}

			isHeadline := (articleNum%15 == 0) && status == "published"
			isBreaking := (articleNum%25 == 0) && status == "published"
			viewCount := r.Int63n(25000) + 100

			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			valueArgs = append(valueArgs,
				title,
				slug,
				excerpt,
				body,
				mediaID,
				cat.id,
				author.id,
				1, // created_by
				status,
				isHeadline,
				isBreaking,
				viewCount,
				pubAt,
				title,
				excerpt,
			)
		}

		stmt := fmt.Sprintf(`INSERT INTO articles (
			title, slug, excerpt, body, cover_media_id, category_id, author_id, created_by,
			status, is_headline, is_breaking, view_count, published_at, seo_title, seo_description
		) VALUES %s;`, strings.Join(valueStrings, ","))

		res, err := db.Exec(stmt, valueArgs...)
		if err != nil {
			log.Fatalf("failed inserting article batch: %v", err)
		}

		lastID, _ := res.LastInsertId()
		rowsAff, _ := res.RowsAffected()
		for id := lastID; id < lastID+rowsAff; id++ {
			articleIDs = append(articleIDs, id)
		}
	}

	log.Printf("==> Successfully inserted %d articles into database.", len(articleIDs))

	// 8. Associate Articles with Tags & Daily Views in Batches
	log.Println("==> Associating articles with tags and daily views...")

	// Batch tags
	tagBatchSize := 500
	for i := 0; i < len(articleIDs); i += tagBatchSize {
		end := i + tagBatchSize
		if end > len(articleIDs) {
			end = len(articleIDs)
		}

		var tagValues []string
		var tagArgs []interface{}
		for _, artID := range articleIDs[i:end] {
			// assign 2 random tags per article
			t1 := tagIDs[r.Intn(len(tagIDs))]
			t2 := tagIDs[r.Intn(len(tagIDs))]
			tagValues = append(tagValues, "(?, ?)")
			tagArgs = append(tagArgs, artID, t1)
			if t1 != t2 {
				tagValues = append(tagValues, "(?, ?)")
				tagArgs = append(tagArgs, artID, t2)
			}
		}

		if len(tagValues) > 0 {
			tagStmt := fmt.Sprintf("INSERT IGNORE INTO article_tag (article_id, tag_id) VALUES %s;", strings.Join(tagValues, ","))
			if _, err := db.Exec(tagStmt, tagArgs...); err != nil {
				log.Fatalf("failed inserting tags batch: %v", err)
			}
		}
	}

	// Batch daily views for trending
	viewBatchSize := 500
	for i := 0; i < len(articleIDs); i += viewBatchSize {
		end := i + viewBatchSize
		if end > len(articleIDs) {
			end = len(articleIDs)
		}

		var viewValues []string
		var viewArgs []interface{}
		for _, artID := range articleIDs[i:end] {
			today := now.Format("2006-01-02")
			yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
			todayViews := r.Intn(500) + 10
			yesterdayViews := r.Intn(400) + 5

			viewValues = append(viewValues, "(?, ?, ?)", "(?, ?, ?)")
			viewArgs = append(viewArgs, artID, today, todayViews, artID, yesterday, yesterdayViews)
		}

		if len(viewValues) > 0 {
			viewStmt := fmt.Sprintf("INSERT INTO article_views_daily (article_id, view_date, views) VALUES %s ON DUPLICATE KEY UPDATE views=VALUES(views);", strings.Join(viewValues, ","))
			if _, err := db.Exec(viewStmt, viewArgs...); err != nil {
				log.Fatalf("failed inserting views batch: %v", err)
			}
		}
	}

	log.Printf("==> Database seeding completed successfully in %s!", time.Since(start))
}
