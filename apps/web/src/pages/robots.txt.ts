import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ url }) => {
  const origin = url.origin;
  const content = `User-agent: *
Allow: /

Sitemap: ${origin}/sitemap.xml
Sitemap: ${origin}/news-sitemap.xml
`.trim();

  return new Response(content, {
    headers: {
      'Content-Type': 'text/plain; charset=utf-8',
      'Cache-Control': 'public, max-age=86400',
    },
  });
};
