export interface ApiResponse<T> {
  success: boolean
  data: T
  error?: {
    code: string
    message: string
    details?: unknown
  }
}

export interface PaginatedMeta {
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface PaginatedResponse<T> {
  success: boolean
  data: T[]
  meta: PaginatedMeta
}

export interface UserResponse {
  id: number
  role_id: number
  role_name: string
  name: string
  email: string
  is_active: boolean
  created_at: string
}

export interface AuthResponse {
  access_token: string
  expires_in: number
  user: UserResponse
}

export interface CategoryResponse {
  id: number
  name: string
  slug: string
  description: string
  parent_id?: number | null
  sort_order: number
  is_nav: boolean
  created_at: string
  updated_at: string
}

export interface CategoryRequest {
  name: string
  slug?: string
  description?: string
  parent_id?: number | null
  sort_order?: number
  is_nav?: boolean
}

export interface TagResponse {
  id: number
  name: string
  slug: string
  created_at: string
  updated_at: string
}

export interface TagRequest {
  name: string
  slug?: string
}

export interface AuthorResponse {
  id: number
  user_id?: number | null
  name: string
  slug: string
  bio: string
  avatar_media_id?: number | null
  social_links?: Record<string, string>
  created_at: string
  updated_at: string
}

export interface AuthorRequest {
  user_id?: number | null
  name: string
  slug?: string
  bio?: string
  avatar_media_id?: number | null
  social_links?: Record<string, string>
}

export interface MediaResponse {
  id: number
  filename: string
  file_path: string
  mime_type: string
  size_bytes: number
  width: number
  height: number
  alt_text: string
  caption: string
  photographer: string
  variants: Record<string, string>
  created_at: string
  updated_at: string
}

export type ArticleStatus = 'draft' | 'review' | 'scheduled' | 'published'

export interface ArticleResponse {
  id: number
  title: string
  slug: string
  excerpt: string
  body: string
  cover_media_id?: number | null
  cover_url?: string
  category_id: number
  category_name?: string
  category_slug?: string
  author_id: number
  author_name?: string
  author_slug?: string
  created_by: number
  status: ArticleStatus
  is_headline: boolean
  is_breaking: boolean
  view_count: number
  published_at?: string | null
  seo_title?: string
  seo_description?: string
  created_at: string
  updated_at: string
  tags?: TagResponse[]
  revisions_count?: number
}

export interface ArticleUpsertRequest {
  title: string
  slug?: string
  excerpt: string
  body: string
  cover_media_id?: number | null
  category_id: number
  author_id: number
  status: ArticleStatus
  is_headline?: boolean
  is_breaking?: boolean
  published_at?: string | null
  seo_title?: string
  seo_description?: string
  tag_ids?: number[]
}

export interface ArticleRevisionResponse {
  id: number
  article_id: number
  user_id: number
  title: string
  slug: string
  excerpt: string
  body: string
  created_at: string
}

export interface RedirectResponse {
  id: number
  from_path: string
  to_path: string
  http_status: number
  hit_count: number
  created_at: string
  updated_at: string
}

export interface RedirectRequest {
  from_path: string
  to_path: string
  http_status: number
}

export interface DashboardStatsResponse {
  total_articles: number
  published_articles: number
  draft_articles: number
  scheduled_articles: number
  total_views: number
  total_categories: number
  total_authors: number
  total_media: number
}
