import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/auth-store'
import {
  ApiResponse,
  AuthResponse,
  PaginatedResponse,
  UserResponse,
  CategoryResponse,
  CategoryRequest,
  TagResponse,
  TagRequest,
  AuthorResponse,
  AuthorRequest,
  MediaResponse,
  ArticleResponse,
  ArticleUpsertRequest,
  ArticleRevisionResponse,
  RedirectResponse,
  RedirectRequest,
  DashboardStatsResponse,
} from '@/types/api'

export const apiClient = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Attach access token to requests
apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = useAuthStore.getState().auth.accessToken
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Mutex / Queue for token refresh
let isRefreshing = false
let failedQueue: Array<{
  resolve: (value?: unknown) => void
  reject: (reason?: unknown) => void
}> = []

const processQueue = (error: AxiosError | null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve()
    }
  })
  failedQueue = []
}

// Refresh token interceptor on 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean
    }

    if (!originalRequest) {
      return Promise.reject(error)
    }

    // Ignore refresh or login failure loops
    if (
      originalRequest.url?.includes('/auth/login') ||
      originalRequest.url?.includes('/auth/refresh')
    ) {
      return Promise.reject(error)
    }

    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then(() => apiClient(originalRequest))
          .catch((err) => Promise.reject(err))
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const { data } = await axios.post<ApiResponse<AuthResponse>>(
          '/api/v1/auth/refresh',
          {},
          { withCredentials: true }
        )

        const newAccessToken = data.data.access_token
        useAuthStore.getState().auth.setAccessToken(newAccessToken)

        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
        }

        processQueue(null)
        return apiClient(originalRequest)
      } catch (refreshErr) {
        processQueue(refreshErr as AxiosError)
        useAuthStore.getState().auth.reset()
        if (typeof window !== 'undefined' && !window.location.pathname.includes('/sign-in')) {
          window.location.href = `/sign-in?redirect=${encodeURIComponent(window.location.pathname)}`
        }
        return Promise.reject(refreshErr)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  }
)

// =========================================================================
// API Services
// =========================================================================

export const authApi = {
  login: async (email: string, password: string): Promise<AuthResponse> => {
    const res = await apiClient.post<ApiResponse<AuthResponse>>('/auth/login', {
      email,
      password,
    })
    return res.data.data
  },
  refresh: async (): Promise<AuthResponse> => {
    const res = await apiClient.post<ApiResponse<AuthResponse>>('/auth/refresh')
    return res.data.data
  },
  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout')
    useAuthStore.getState().auth.reset()
  },
  getMe: async (): Promise<UserResponse> => {
    const res = await apiClient.get<ApiResponse<UserResponse>>('/auth/me')
    return res.data.data
  },
}

export const dashboardApi = {
  getStats: async (): Promise<DashboardStatsResponse> => {
    const res = await apiClient.get<ApiResponse<DashboardStatsResponse>>(
      '/admin/dashboard/stats'
    )
    return res.data.data
  },
}

export interface ArticleFilterParams {
  page?: number
  page_size?: number
  status?: string
  category_id?: number
  search?: string
}

export const articleApi = {
  list: async (params?: ArticleFilterParams): Promise<PaginatedResponse<ArticleResponse>> => {
    const res = await apiClient.get<PaginatedResponse<ArticleResponse>>(
      '/admin/articles',
      { params }
    )
    return res.data
  },
  get: async (id: number): Promise<ArticleResponse> => {
    const res = await apiClient.get<ApiResponse<ArticleResponse>>(
      `/admin/articles/${id}`
    )
    return res.data.data
  },
  create: async (payload: ArticleUpsertRequest): Promise<ArticleResponse> => {
    const res = await apiClient.post<ApiResponse<ArticleResponse>>(
      '/admin/articles',
      payload
    )
    return res.data.data
  },
  update: async (
    id: number,
    payload: ArticleUpsertRequest
  ): Promise<ArticleResponse> => {
    const res = await apiClient.put<ApiResponse<ArticleResponse>>(
      `/admin/articles/${id}`,
      payload
    )
    return res.data.data
  },
  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/admin/articles/${id}`)
  },
  listRevisions: async (id: number): Promise<ArticleRevisionResponse[]> => {
    const res = await apiClient.get<ApiResponse<ArticleRevisionResponse[]>>(
      `/admin/articles/${id}/revisions`
    )
    return res.data.data
  },
  restoreRevision: async (id: number, revisionId: number): Promise<void> => {
    await apiClient.post(`/admin/articles/${id}/restore/${revisionId}`)
  },
  getPreviewToken: async (id: number): Promise<string> => {
    const res = await apiClient.post<ApiResponse<{ preview_token: string }>>(
      `/admin/articles/${id}/preview-token`
    )
    return res.data.data.preview_token
  },
}

export const mediaApi = {
  list: async (page = 1, pageSize = 30): Promise<PaginatedResponse<MediaResponse>> => {
    const res = await apiClient.get<PaginatedResponse<MediaResponse>>(
      '/admin/media',
      {
        params: { page, page_size: pageSize },
      }
    )
    return res.data
  },
  upload: async (file: File, caption?: string, altText?: string): Promise<MediaResponse> => {
    const formData = new FormData()
    formData.append('file', file)
    if (caption) formData.append('caption', caption)
    if (altText) formData.append('alt_text', altText)

    const res = await apiClient.post<ApiResponse<MediaResponse>>(
      '/admin/media/upload',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )
    return res.data.data
  },
  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/admin/media/${id}`)
  },
}

export const categoryApi = {
  list: async (): Promise<CategoryResponse[]> => {
    const res = await apiClient.get<ApiResponse<CategoryResponse[]>>(
      '/admin/categories'
    )
    return res.data.data
  },
  create: async (data: CategoryRequest): Promise<CategoryResponse> => {
    const res = await apiClient.post<ApiResponse<CategoryResponse>>(
      '/admin/categories',
      data
    )
    return res.data.data
  },
  update: async (id: number, data: CategoryRequest): Promise<CategoryResponse> => {
    const res = await apiClient.put<ApiResponse<CategoryResponse>>(
      `/admin/categories/${id}`,
      data
    )
    return res.data.data
  },
  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/admin/categories/${id}`)
  },
}

export const tagApi = {
  list: async (): Promise<TagResponse[]> => {
    const res = await apiClient.get<ApiResponse<TagResponse[]>>('/admin/tags')
    return res.data.data
  },
  create: async (data: TagRequest): Promise<TagResponse> => {
    const res = await apiClient.post<ApiResponse<TagResponse>>(
      '/admin/tags',
      data
    )
    return res.data.data
  },
  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/admin/tags/${id}`)
  },
}

export const authorApi = {
  list: async (): Promise<AuthorResponse[]> => {
    const res = await apiClient.get<ApiResponse<AuthorResponse[]>>(
      '/admin/authors'
    )
    return res.data.data
  },
  create: async (data: AuthorRequest): Promise<AuthorResponse> => {
    const res = await apiClient.post<ApiResponse<AuthorResponse>>(
      '/admin/authors',
      data
    )
    return res.data.data
  },
  update: async (id: number, data: AuthorRequest): Promise<AuthorResponse> => {
    const res = await apiClient.put<ApiResponse<AuthorResponse>>(
      `/admin/authors/${id}`,
      data
    )
    return res.data.data
  },
}

export const redirectApi = {
  list: async (): Promise<RedirectResponse[]> => {
    const res = await apiClient.get<ApiResponse<RedirectResponse[]>>(
      '/admin/redirects'
    )
    return res.data.data
  },
  create: async (data: RedirectRequest): Promise<RedirectResponse> => {
    const res = await apiClient.post<ApiResponse<RedirectResponse>>(
      '/admin/redirects',
      data
    )
    return res.data.data
  },
}
