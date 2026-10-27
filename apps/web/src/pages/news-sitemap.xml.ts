import type { APIRoute } from 'astro';
import { getArticles } from '../lib/api';

export const GET: APIRoute = async ({ url }) => {
  const origin = url.origin;
  const feedData = await getArticles(undefined, 50);
  const articles = feedData?.articles || [];

  // Google News requires articles published within the last 48 hours
  const twoDaysAgo = Date.now() - 48 * 60 * 60 * 1000;
  const newsArticles = articles.filter((a) => {
    if (!a.published_at) return false;
    return new Date(a.published_at).getTime() >= twoDaysAgo;
  });

  const newsXml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
        xmlns:news="http://www.google.com/schemas/sitemap-news/0.9">
  ${newsArticles.map((a) => `
  <url>
    <loc>${origin}/${a.slug}</loc>
    <news:news>
      <news:publication>
        <news:name>Tailnews</news:name>
        <news:language>id</news:language>
      </news:publication>
      <news:publication_date>${a.published_at ? new Date(a.published_at).toISOString() : new Date().toISOString()}</news:publication_date>
      <news:title><![CDATA[${a.title}]]></news:title>
    </news:news>
  </url>`).join('')}
</urlset>`.trim();

  return new Response(newsXml, {
    headers: {
      'Content-Type': 'application/xml; charset=utf-8',
      'Cache-Control': 'public, max-age=180, s-maxage=300',
    },
  });
};
