export interface Article {
  id: number;
  title: string;
  slug: string;
  excerpt: string;
  body?: string;
  cover_media_id?: number | null;
  cover_url?: string;
  category_id: number;
  category_name?: string;
  category_slug?: string;
  author_id: number;
  author_name?: string;
  author_slug?: string;
  is_headline: boolean;
  is_breaking: boolean;
  view_count: number;
  published_at: string | null;
  seo_title?: string;
  seo_description?: string;
  created_at: string;
  updated_at: string;
  tags?: Tag[];
}

export interface Category {
  id: number;
  name: string;
  slug: string;
  description?: string;
  parent_id?: number | null;
  sort_order: number;
  is_nav: boolean;
}

export interface Tag {
  id: number;
  name: string;
  slug: string;
}

export interface Author {
  id: number;
  name: string;
  slug: string;
  bio?: string;
  avatar_media_id?: number | null;
  avatar_url?: string;
  social_links?: Record<string, string>;
}

export interface HomepageData {
  headlines: Article[];
  breaking: Article[];
  latest: Article[];
  next_cursor?: string;
  trending: Article[];
  categories: Category[];
}

export interface CursorArticlesData {
  articles: Article[];
  next_cursor?: string;
  has_more: boolean;
}

export interface AuthorProfileData {
  author: Author;
  articles: Article[];
  next_cursor?: string;
  has_more: boolean;
}

export interface PageData {
  id: number;
  title: string;
  slug: string;
  content: string;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: {
    code: string;
    message: string;
  };
  meta?: Record<string, any>;
}

const getApiBaseUrl = (): string => {
  if (typeof process !== 'undefined' && process.env) {
    if (process.env.API_INTERNAL_URL) return process.env.API_INTERNAL_URL;
    if (process.env.API_URL) return process.env.API_URL;
  }
  return 'http://127.0.0.1:8080/api/v1';
};

async function fetchFromApi<T>(endpoint: string, options: RequestInit = {}): Promise<T | null> {
  const baseUrl = getApiBaseUrl();
  const url = `${baseUrl}${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`;

  try {
    const res = await fetch(url, {
      ...options,
      headers: {
        'Accept': 'application/json',
        'User-Agent': 'NewsPortal-Web-SSR/1.0',
        ...(options.headers || {}),
      },
    });

    if (!res.ok) {
      if (res.status === 404) {
        return null;
      }
      const errText = await res.text();
      console.error(`[API Error ${res.status}] ${url}: ${errText}`);
      return null;
    }

    const json: ApiResponse<T> = await res.json();
    if (!json.success) {
      console.error(`[API Fail] ${url}:`, json.error);
      return null;
    }

    return json.data;
  } catch (err) {
    console.error(`[API Connection Failed] ${url}:`, err);
    return null;
  }
}

// 1. Homepage Aggregated Feed
export async function getHomepage(): Promise<HomepageData | null> {
  return fetchFromApi<HomepageData>('/homepage');
}

// 2. Keyset Cursor Paginated Articles Feed
export async function getArticles(cursor?: string, limit: number = 12): Promise<CursorArticlesData | null> {
  const params = new URLSearchParams();
  if (cursor) params.set('cursor', cursor);
  params.set('limit', String(limit));
  return fetchFromApi<CursorArticlesData>(`/articles?${params.toString()}`);
}

// 3. Article Detail by Slug (with optional preview token)
export async function getArticleBySlug(slug: string, token?: string): Promise<Article | null> {
  const params = new URLSearchParams();
  if (token) params.set('token', token);
  const q = params.toString() ? `?${params.toString()}` : '';
  return fetchFromApi<Article>(`/articles/${encodeURIComponent(slug)}${q}`);
}

// 4. Categories
export async function getCategories(): Promise<Category[]> {
  const cats = await fetchFromApi<Category[]>('/categories');
  return cats || [];
}

// 5. Articles by Category Slug
export async function getArticlesByCategory(
  catSlug: string,
  cursor?: string,
  limit: number = 12,
): Promise<CursorArticlesData | null> {
  const params = new URLSearchParams();
  if (cursor) params.set('cursor', cursor);
  params.set('limit', String(limit));
  return fetchFromApi<CursorArticlesData>(`/categories/${encodeURIComponent(catSlug)}/articles?${params.toString()}`);
}

// 6. Articles by Tag Slug
export async function getArticlesByTag(
  tagSlug: string,
  cursor?: string,
  limit: number = 12,
): Promise<CursorArticlesData | null> {
  const params = new URLSearchParams();
  if (cursor) params.set('cursor', cursor);
  params.set('limit', String(limit));
  return fetchFromApi<CursorArticlesData>(`/tags/${encodeURIComponent(tagSlug)}/articles?${params.toString()}`);
}

// 7. Author Profile & Feed
export async function getAuthor(
  authorSlug: string,
  cursor?: string,
  limit: number = 12,
): Promise<AuthorProfileData | null> {
  const params = new URLSearchParams();
  if (cursor) params.set('cursor', cursor);
  params.set('limit', String(limit));
  return fetchFromApi<AuthorProfileData>(`/authors/${encodeURIComponent(authorSlug)}?${params.toString()}`);
}

// 8. Search Articles
export async function searchArticles(query: string, limit: number = 15): Promise<Article[]> {
  const params = new URLSearchParams({ q: query, limit: String(limit) });
  const results = await fetchFromApi<Article[]>(`/search?${params.toString()}`);
  return results || [];
}

// 9. Trending Articles
export async function getTrending(window: 'today' | 'week' = 'today', limit: number = 10): Promise<Article[]> {
  const params = new URLSearchParams({ window, limit: String(limit) });
  const results = await fetchFromApi<Article[]>(`/trending?${params.toString()}`);
  return results || [];
}

// 10. Static Editorial Page
export async function getPage(slug: string): Promise<PageData | null> {
  return fetchFromApi<PageData>(`/pages/${encodeURIComponent(slug)}`);
}

// Helper: Format Date in Indonesian Locale
export function formatDate(dateStr: string | null | undefined): string {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  return new Intl.DateTimeFormat('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }).format(date);
}

// Helper: Estimate Reading Time
export function estimateReadingTime(content: string): number {
  const wordsPerMinute = 200;
  const wordCount = content.trim().split(/\s+/).length;
  return Math.max(1, Math.ceil(wordCount / wordsPerMinute));
}
