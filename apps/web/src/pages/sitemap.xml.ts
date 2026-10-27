import type { APIRoute } from 'astro';
import { getArticles, getCategories } from '../lib/api';

export const GET: APIRoute = async ({ url }) => {
  const origin = url.origin;
  const [feedData, categories] = await Promise.all([
    getArticles(undefined, 50),
    getCategories(),
  ]);

  const articles = feedData?.articles || [];

  const sitemapXml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <!-- Homepage -->
  <url>
    <loc>${origin}/</loc>
    <changefreq>hourly</changefreq>
    <priority>1.0</priority>
  </url>

  <!-- Categories -->
  ${categories.map((c) => `
  <url>
    <loc>${origin}/category/${c.slug}</loc>
    <changefreq>hourly</changefreq>
    <priority>0.8</priority>
  </url>`).join('')}

  <!-- Latest Articles -->
  ${articles.map((a) => `
  <url>
    <loc>${origin}/${a.slug}</loc>
    <lastmod>${a.published_at ? new Date(a.published_at).toISOString() : new Date().toISOString()}</lastmod>
    <changefreq>daily</changefreq>
    <priority>0.7</priority>
  </url>`).join('')}
</urlset>`.trim();

  return new Response(sitemapXml, {
    headers: {
      'Content-Type': 'application/xml; charset=utf-8',
      'Cache-Control': 'public, max-age=300, s-maxage=600',
    },
  });
};
