import { useState, useEffect } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import {
  FilePlus,
  Search,
  MoreHorizontal,
  Edit,
  History,
  Eye,
  Trash2,
  Loader2,
  ChevronLeft,
  ChevronRight,
  TrendingUp,
} from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { RevisionsDialog } from './components/revisions-dialog'
import { articleApi, categoryApi } from '@/lib/api-client'
import { ArticleResponse, CategoryResponse } from '@/types/api'

export function ArticlesFeature() {
  const navigate = useNavigate()
  const [articles, setArticles] = useState<ArticleResponse[]>([])
  const [categories, setCategories] = useState<CategoryResponse[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(15)
  const [totalPages, setTotalPages] = useState(1)
  const [totalCount, setTotalCount] = useState(0)

  // Filters
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [categoryFilter, setCategoryFilter] = useState<string>('all')
  const [search, setSearch] = useState('')

  // Revisions Modal State
  const [revisionArticle, setRevisionArticle] = useState<ArticleResponse | null>(null)

  const loadArticles = async () => {
    setIsLoading(true)
    try {
      const params: Record<string, unknown> = {
        page,
        page_size: pageSize,
      }
      if (statusFilter !== 'all') params.status = statusFilter
      if (categoryFilter !== 'all') params.category_id = Number(categoryFilter)
      if (search.trim()) params.search = search.trim()

      const res = await articleApi.list(params)
      setArticles(res.data || [])
      setTotalPages(res.meta?.total_pages ?? 1)
      setTotalCount(res.meta?.total ?? 0)
    } catch {
      toast.error('Failed to load articles list')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    categoryApi.list().then(setCategories).catch(() => {})
  }, [])

  useEffect(() => {
    loadArticles()
  }, [page, statusFilter, categoryFilter])

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setPage(1)
    loadArticles()
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this article?')) return
    try {
      await articleApi.delete(id)
      toast.success('Article deleted successfully')
      loadArticles()
    } catch {
      toast.error('Failed to delete article')
    }
  }

  const handleLivePreview = async (art: ArticleResponse) => {
    try {
      const token = await articleApi.getPreviewToken(art.id)
      const previewUrl = `http://127.0.0.1:4321/${art.slug}?preview=${token}`
      window.open(previewUrl, '_blank')
    } catch {
      toast.error('Failed to generate preview token')
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'published':
        return (
          <Badge variant='outline' className='bg-green-500/10 text-green-600 border-green-500/20'>
            Published
          </Badge>
        )
      case 'draft':
        return (
          <Badge variant='outline' className='bg-muted text-muted-foreground'>
            Draft
          </Badge>
        )
      case 'review':
        return (
          <Badge variant='outline' className='bg-amber-500/10 text-amber-600 border-amber-500/20'>
            Review
          </Badge>
        )
      case 'scheduled':
        return (
          <Badge variant='outline' className='bg-blue-500/10 text-blue-600 border-blue-500/20'>
            Scheduled
          </Badge>
        )
      default:
        return <Badge variant='outline'>{status}</Badge>
    }
  }

  return (
    <div className='p-6 space-y-6'>
      {/* Header */}
      <div className='flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>Articles Management</h1>
          <p className='text-sm text-muted-foreground'>
            Total {(totalCount ?? 0).toLocaleString()} articles across all categories and workflows.
          </p>
        </div>
        <Button onClick={() => navigate({ to: '/articles/new' })}>
          <FilePlus className='h-4 w-4 me-2' />
          Write New Article
        </Button>
      </div>

      {/* Filter and Search Bar */}
      <div className='flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3'>
        <div className='flex flex-wrap items-center gap-2'>
          {['all', 'published', 'draft', 'review', 'scheduled'].map((s) => (
            <Button
              key={s}
              size='sm'
              variant={statusFilter === s ? 'default' : 'outline'}
              className='capitalize text-xs h-8'
              onClick={() => {
                setStatusFilter(s)
                setPage(1)
              }}
            >
              {s}
            </Button>
          ))}
        </div>

        <div className='flex items-center gap-2'>
          <Select
            value={categoryFilter}
            onValueChange={(val) => {
              setCategoryFilter(val)
              setPage(1)
            }}
          >
            <SelectTrigger className='w-[160px] h-8 text-xs'>
              <SelectValue placeholder='All Categories' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='all'>All Categories</SelectItem>
              {categories.map((c) => (
                <SelectItem key={c.id} value={String(c.id)}>
                  {c.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <form onSubmit={handleSearchSubmit} className='relative w-[220px]'>
            <Search className='absolute left-2.5 top-2 h-3.5 w-3.5 text-muted-foreground' />
            <Input
              placeholder='Search articles...'
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className='pl-8 h-8 text-xs'
            />
          </form>
        </div>
      </div>

      {/* Table */}
      <div className='border rounded-lg overflow-hidden bg-card'>
        <div className='overflow-x-auto'>
          <table className='w-full text-left text-xs'>
            <thead className='bg-muted/50 border-b text-muted-foreground uppercase tracking-wider font-semibold'>
              <tr>
                <th className='p-3 w-16'>Cover</th>
                <th className='p-3'>Headline & Category</th>
                <th className='p-3 w-32'>Author</th>
                <th className='p-3 w-28'>Status</th>
                <th className='p-3 w-24 text-right'>Views</th>
                <th className='p-3 w-32'>Published</th>
                <th className='p-3 w-16 text-center'>Actions</th>
              </tr>
            </thead>
            <tbody className='divide-y divide-border'>
              {isLoading ? (
                <tr>
                  <td colSpan={7} className='p-12 text-center text-muted-foreground'>
                    <Loader2 className='h-6 w-6 animate-spin mx-auto mb-2' />
                    Loading articles...
                  </td>
                </tr>
              ) : articles.length === 0 ? (
                <tr>
                  <td colSpan={7} className='p-12 text-center text-muted-foreground'>
                    No articles found matching the current filters.
                  </td>
                </tr>
              ) : (
                articles.map((art) => (
                  <tr key={art.id} className='hover:bg-muted/30 transition-colors'>
                    <td className='p-3'>
                      <div className='h-10 w-14 rounded overflow-hidden bg-muted border flex items-center justify-center shrink-0'>
                        {art.cover_url ? (
                          <img
                            src={art.cover_url}
                            alt={art.title}
                            className='h-full w-full object-cover'
                            loading='lazy'
                          />
                        ) : (
                          <span className='text-[9px] text-muted-foreground'>No Cover</span>
                        )}
                      </div>
                    </td>

                    <td className='p-3'>
                      <Link
                        to='/articles/$id'
                        params={{ id: String(art.id) }}
                        className='font-semibold text-sm hover:underline line-clamp-1 text-foreground'
                      >
                        {art.title}
                      </Link>
                      <div className='flex items-center gap-2 mt-1'>
                        <span className='text-[10px] bg-muted px-1.5 py-0.5 rounded text-muted-foreground font-medium'>
                          {art.category_name || 'General'}
                        </span>
                        {art.is_headline && (
                          <Badge variant='secondary' className='text-[9px] px-1 py-0 h-4'>
                            Headline
                          </Badge>
                        )}
                        {art.is_breaking && (
                          <Badge variant='destructive' className='text-[9px] px-1 py-0 h-4'>
                            Breaking
                          </Badge>
                        )}
                      </div>
                    </td>

                    <td className='p-3 text-muted-foreground font-medium truncate'>
                      {art.author_name || 'Staff'}
                    </td>

                    <td className='p-3'>{getStatusBadge(art.status)}</td>

                    <td className='p-3 text-right font-mono font-medium'>
                      <span className='inline-flex items-center gap-1 text-muted-foreground'>
                        <TrendingUp className='h-3 w-3 text-emerald-500' />
                        {(art.view_count ?? 0).toLocaleString()}
                      </span>
                    </td>

                    <td className='p-3 text-muted-foreground'>
                      {art.published_at
                        ? new Date(art.published_at).toLocaleDateString()
                        : '—'}
                    </td>

                    <td className='p-3 text-center'>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant='ghost' size='sm' className='h-8 w-8 p-0'>
                            <MoreHorizontal className='h-4 w-4' />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align='end' className='w-44'>
                          <DropdownMenuItem
                            onClick={() => navigate({ to: `/articles/${art.id}` })}
                          >
                            <Edit className='h-3.5 w-3.5 me-2' />
                            Edit Article
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleLivePreview(art)}>
                            <Eye className='h-3.5 w-3.5 me-2' />
                            Preview Live
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => setRevisionArticle(art)}>
                            <History className='h-3.5 w-3.5 me-2' />
                            View Revisions
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            variant='destructive'
                            onClick={() => handleDelete(art.id)}
                          >
                            <Trash2 className='h-3.5 w-3.5 me-2' />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Bar */}
        <div className='flex items-center justify-between p-3 border-t bg-muted/20 text-xs text-muted-foreground'>
          <span>
            Page {page} of {totalPages} ({totalCount} total articles)
          </span>
          <div className='flex items-center gap-2'>
            <Button
              variant='outline'
              size='sm'
              disabled={page <= 1 || isLoading}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              className='h-7 px-2'
            >
              <ChevronLeft className='h-3.5 w-3.5 me-1' /> Previous
            </Button>
            <Button
              variant='outline'
              size='sm'
              disabled={page >= totalPages || isLoading}
              onClick={() => setPage((p) => p + 1)}
              className='h-7 px-2'
            >
              Next <ChevronRight className='h-3.5 w-3.5 ms-1' />
            </Button>
          </div>
        </div>
      </div>

      {/* Revision Dialog */}
      {revisionArticle && (
        <RevisionsDialog
          articleId={revisionArticle.id}
          articleTitle={revisionArticle.title}
          open={!!revisionArticle}
          onOpenChange={(open) => !open && setRevisionArticle(null)}
          onRestored={loadArticles}
        />
      )}
    </div>
  )
}
