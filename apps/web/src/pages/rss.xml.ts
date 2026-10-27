import type { APIRoute } from 'astro';
import { getArticles } from '../lib/api';

export const GET: APIRoute = async ({ url }) => {
  const origin = url.origin;
  const feedData = await getArticles(undefined, 30);
  const articles = feedData?.articles || [];

  const rssXml = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Tailnews - Portal Berita Modern</title>
    <link>${origin}</link>
    <description>Sindikasi berita terkini, aktual, dan terpercaya dari redaksi Tailnews.</description>
    <language>id-ID</language>
    <lastBuildDate>${new Date().toUTCString()}</lastBuildDate>
    <atom:link href="${origin}/rss.xml" rel="self" type="application/rss+xml"/>
    ${articles.map((a) => `
    <item>
      <title><![CDATA[${a.title}]]></title>
      <link>${origin}/${a.slug}</link>
      <guid isPermaLink="true">${origin}/${a.slug}</guid>
      <pubDate>${a.published_at ? new Date(a.published_at).toUTCString() : new Date().toUTCString()}</pubDate>
      <description><![CDATA[${a.excerpt}]]></description>
      ${a.category_name ? `<category><![CDATA[${a.category_name}]]></category>` : ''}
    </item>`).join('')}
  </channel>
</rss>`.trim();

  return new Response(rssXml, {
    headers: {
      'Content-Type': 'application/rss+xml; charset=utf-8',
      'Cache-Control': 'public, max-age=300, s-maxage=600',
    },
  });
};
