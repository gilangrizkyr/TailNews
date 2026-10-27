import type { APIRoute } from 'astro';

export const ALL: APIRoute = async ({ request, params }) => {
  const path = params.path || '';
  const url = new URL(request.url);
  const targetUrl = `http://127.0.0.1:8080/api/v1/${path}${url.search}`;

  try {
    const fetchOptions: RequestInit = {
      method: request.method,
      headers: {
        'Content-Type': request.headers.get('content-type') || 'application/json',
        'Accept': 'application/json',
        'User-Agent': request.headers.get('user-agent') || 'NewsPortal-SSR-Proxy',
      },
    };

    if (!['GET', 'HEAD'].includes(request.method)) {
      const bodyText = await request.text();
      if (bodyText) {
        fetchOptions.body = bodyText;
      }
    }

    const apiRes = await fetch(targetUrl, fetchOptions);
    const data = await apiRes.text();

    return new Response(data, {
      status: apiRes.status,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  } catch (err: any) {
    return new Response(
      JSON.stringify({
        success: false,
        error: {
          code: 'PROXY_ERROR',
          message: err?.message || 'Failed to reach API server',
        },
      }),
      {
        status: 502,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  }
};
