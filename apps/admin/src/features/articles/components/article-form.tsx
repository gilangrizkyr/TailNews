import { useState, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  Save,
  Eye,
  Loader2,
  Image as ImageIcon,
  ArrowLeft,
  AlertCircle,
} from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { TipTapEditor } from './editor/tiptap-editor'
import { MediaPickerDialog } from '@/features/media/components/media-picker-dialog'
import {
  articleApi,
  categoryApi,
  authorApi,
  tagApi,
} from '@/lib/api-client'
import {
  ArticleResponse,
  ArticleStatus,
  ArticleUpsertRequest,
  CategoryResponse,
  AuthorResponse,
  TagResponse,
  MediaResponse,
} from '@/types/api'

interface ArticleFormProps {
  initialArticle?: ArticleResponse
  isEdit?: boolean
}

export function ArticleForm({ initialArticle, isEdit = false }: ArticleFormProps) {
  const navigate = useNavigate()

  // Form State
  const [title, setTitle] = useState(initialArticle?.title || '')
  const [slug, setSlug] = useState(initialArticle?.slug || '')
  const [originalSlug] = useState(initialArticle?.slug || '')
  const [excerpt, setExcerpt] = useState(initialArticle?.excerpt || '')
  const [body, setBody] = useState(initialArticle?.body || '')
  const [categoryId, setCategoryId] = useState<number | undefined>(
    initialArticle?.category_id
  )
  const [authorId, setAuthorId] = useState<number | undefined>(
    initialArticle?.author_id
  )
  const [status, setStatus] = useState<ArticleStatus>(
    initialArticle?.status || 'draft'
  )
  const [isHeadline, setIsHeadline] = useState(initialArticle?.is_headline || false)
  const [isBreaking, setIsBreaking] = useState(initialArticle?.is_breaking || false)
  const [publishedAt, setPublishedAt] = useState<string>(
    initialArticle?.published_at
      ? new Date(initialArticle.published_at).toISOString().slice(0, 16)
      : ''
  )
  const [seoTitle, setSeoTitle] = useState(initialArticle?.seo_title || '')
  const [seoDescription, setSeoDescription] = useState(
    initialArticle?.seo_description || ''
  )
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>(
    initialArticle?.tags?.map((t) => t.id) || []
  )

  // Media Cover
  const [coverMedia, setCoverMedia] = useState<MediaResponse | null>(null)
  const [isMediaPickerOpen, setIsMediaPickerOpen] = useState(false)
  const [isEditorMediaPickerOpen, setIsEditorMediaPickerOpen] = useState(false)

  // Options State
  const [categories, setCategories] = useState<CategoryResponse[]>([])
  const [authors, setAuthors] = useState<AuthorResponse[]>([])
  const [availableTags, setAvailableTags] = useState<TagResponse[]>([])
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isPreviewing, setIsPreviewing] = useState(false)

  useEffect(() => {
    const fetchOptions = async () => {
      try {
        const [cats, auths, tags] = await Promise.all([
          categoryApi.list(),
          authorApi.list(),
          tagApi.list(),
        ])
        setCategories(cats)
        setAuthors(auths)
        setAvailableTags(tags)

        if (!isEdit) {
          if (cats.length > 0) setCategoryId(cats[0].id)
          if (auths.length > 0) setAuthorId(auths[0].id)
        }
      } catch {
        toast.error('Failed to load editorial options')
      }
    }
    fetchOptions()
  }, [isEdit])

  // Slug auto-generation on Title change for new articles
  const handleTitleChange = (newTitle: string) => {
    setTitle(newTitle)
    if (!isEdit) {
      const generatedSlug = newTitle
        .toLowerCase()
        .replace(/[^\w\s-]/g, '')
        .trim()
        .replace(/\s+/g, '-')
      setSlug(generatedSlug)
    }
  }

  const handleToggleTag = (tagId: number) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    )
  }

  const handlePreview = async () => {
    if (!initialArticle?.id) {
      toast.info('Please save the article first before generating a live preview token')
      return
    }

    setIsPreviewing(true)
    try {
      const token = await articleApi.getPreviewToken(initialArticle.id)
      const previewUrl = `http://127.0.0.1:4321/${initialArticle.slug}?preview=${token}`
      window.open(previewUrl, '_blank')
    } catch {
      toast.error('Failed to generate preview token')
    } finally {
      setIsPreviewing(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!title.trim()) {
      toast.error('Article title is required')
      return
    }
    if (!excerpt.trim()) {
      toast.error('Article excerpt summary is required')
      return
    }
    if (!body.trim()) {
      toast.error('Article body content cannot be empty')
      return
    }
    if (!categoryId) {
      toast.error('Please select a category')
      return
    }
    if (!authorId) {
      toast.error('Please assign an author')
      return
    }

    setIsSubmitting(true)
    try {
      const payload: ArticleUpsertRequest = {
        title: title.trim(),
        slug: slug.trim() || undefined,
        excerpt: excerpt.trim(),
        body: body.trim(),
        category_id: categoryId,
        author_id: authorId,
        status,
        is_headline: isHeadline,
        is_breaking: isBreaking,
        published_at: publishedAt ? new Date(publishedAt).toISOString() : null,
        seo_title: seoTitle.trim() || undefined,
        seo_description: seoDescription.trim() || undefined,
        tag_ids: selectedTagIds,
        cover_media_id: coverMedia?.id || initialArticle?.cover_media_id || null,
      }

      if (isEdit && initialArticle) {
        await articleApi.update(initialArticle.id, payload)
        toast.success('Article updated successfully!')
      } else {
        const created = await articleApi.create(payload)
        toast.success('Article created successfully!')
        navigate({ to: `/articles/${created.id}` })
      }
    } catch (err: unknown) {
      const errMsg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || 'Failed to save article'
      toast.error(errMsg)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className='p-6 space-y-6 max-w-6xl mx-auto'>
      {/* Top Action Bar */}
      <div className='flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-4 border-b'>
        <div className='flex items-center gap-3'>
          <Button
            type='button'
            variant='outline'
            size='sm'
            onClick={() => navigate({ to: '/articles' })}
          >
            <ArrowLeft className='h-4 w-4 me-1.5' /> Back
          </Button>
          <div>
            <h1 className='text-xl font-bold tracking-tight'>
              {isEdit ? 'Edit Article' : 'Write New Article'}
            </h1>
            <p className='text-xs text-muted-foreground'>
              {isEdit
                ? `Editing ID #${initialArticle?.id} • Current status: ${initialArticle?.status}`
                : 'Draft and publish high-quality news content.'}
            </p>
          </div>
        </div>

        <div className='flex items-center gap-2'>
          {isEdit && (
            <Button
              type='button'
              variant='outline'
              size='sm'
              disabled={isPreviewing}
              onClick={handlePreview}
            >
              {isPreviewing ? (
                <Loader2 className='h-4 w-4 animate-spin me-1.5' />
              ) : (
                <Eye className='h-4 w-4 me-1.5' />
              )}
              Preview Live
            </Button>
          )}

          <Button type='submit' size='sm' disabled={isSubmitting}>
            {isSubmitting ? (
              <Loader2 className='h-4 w-4 animate-spin me-1.5' />
            ) : (
              <Save className='h-4 w-4 me-1.5' />
            )}
            {isEdit ? 'Save Changes' : 'Publish / Save'}
          </Button>
        </div>
      </div>

      {/* Main Grid Layout */}
      <div className='grid grid-cols-1 lg:grid-cols-3 gap-6'>
        {/* Left 2 Columns: Main Editorial Content */}
        <div className='lg:col-span-2 space-y-5'>
          {/* Title */}
          <div className='space-y-2'>
            <div className='flex justify-between items-center'>
              <Label htmlFor='title' className='font-semibold text-sm'>
                Headline Title <span className='text-red-500'>*</span>
              </Label>
              <span className='text-[11px] text-muted-foreground'>
                {title.length}/255
              </span>
            </div>
            <Input
              id='title'
              placeholder='Enter a captivating news headline...'
              value={title}
              onChange={(e) => handleTitleChange(e.target.value)}
              className='text-lg font-semibold h-11'
              required
            />
          </div>

          {/* Slug */}
          <div className='space-y-2'>
            <div className='flex justify-between items-center'>
              <Label htmlFor='slug' className='text-xs font-semibold text-muted-foreground'>
                URL Slug
              </Label>
              {isEdit && slug !== originalSlug && (
                <span className='text-[11px] text-amber-600 flex items-center gap-1 font-medium'>
                  <AlertCircle className='h-3 w-3' /> Automatic 301 redirect will be created
                </span>
              )}
            </div>
            <div className='flex items-center rounded-md border bg-muted/30 px-3'>
              <span className='text-xs text-muted-foreground font-mono'>/</span>
              <Input
                id='slug'
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                placeholder='slug-of-the-article'
                className='border-0 bg-transparent text-xs font-mono shadow-none focus-visible:ring-0 px-1'
              />
            </div>
          </div>

          {/* Excerpt */}
          <div className='space-y-2'>
            <div className='flex justify-between items-center'>
              <Label htmlFor='excerpt' className='font-semibold text-sm'>
                Lead Paragraph / Excerpt <span className='text-red-500'>*</span>
              </Label>
              <span className='text-[11px] text-muted-foreground'>
                {excerpt.length}/500
              </span>
            </div>
            <Textarea
              id='excerpt'
              placeholder='Brief 1-2 sentence lead summarizing the key news hook...'
              value={excerpt}
              onChange={(e) => setExcerpt(e.target.value)}
              className='min-h-[85px] text-sm leading-relaxed'
              required
            />
          </div>

          {/* TipTap Rich Text Editor */}
          <div className='space-y-2'>
            <Label className='font-semibold text-sm'>
              Article Body Story <span className='text-red-500'>*</span>
            </Label>
            <TipTapEditor
              content={body}
              onChange={setBody}
              placeholder='Write the full news story here...'
              onOpenMediaPicker={() => setIsEditorMediaPickerOpen(true)}
            />
          </div>

          {/* SEO Metadata Box */}
          <div className='p-4 border rounded-lg space-y-4 bg-muted/20'>
            <h3 className='font-semibold text-sm'>SEO & Search Engine Optimization</h3>
            <div className='space-y-2'>
              <Label htmlFor='seoTitle' className='text-xs'>Custom SEO Title</Label>
              <Input
                id='seoTitle'
                placeholder='Defaults to article title if left empty'
                value={seoTitle}
                onChange={(e) => setSeoTitle(e.target.value)}
                className='text-xs'
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='seoDesc' className='text-xs'>Meta Description</Label>
              <Textarea
                id='seoDesc'
                placeholder='Defaults to article excerpt if left empty'
                value={seoDescription}
                onChange={(e) => setSeoDescription(e.target.value)}
                className='text-xs min-h-[60px]'
              />
            </div>
          </div>
        </div>

        {/* Right 1 Column: Editorial Controls & Metadata */}
        <div className='space-y-5'>
          {/* Status & Publication Box */}
          <div className='p-4 border rounded-lg space-y-4 bg-card shadow-xs'>
            <h3 className='font-semibold text-sm border-b pb-2'>Publication Controls</h3>

            <div className='space-y-2'>
              <Label className='text-xs font-semibold'>Status</Label>
              <Select
                value={status}
                onValueChange={(val) => setStatus(val as ArticleStatus)}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue placeholder='Select status' />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='draft'>Draft (In Progress)</SelectItem>
                  <SelectItem value='review'>Review (Pending Editor)</SelectItem>
                  <SelectItem value='scheduled'>Scheduled</SelectItem>
                  <SelectItem value='published'>Published (Live)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-2'>
              <Label htmlFor='publishedAt' className='text-xs font-semibold'>
                Publish Date & Time
              </Label>
              <Input
                id='publishedAt'
                type='datetime-local'
                value={publishedAt}
                onChange={(e) => setPublishedAt(e.target.value)}
                className='text-xs'
              />
            </div>

            <div className='pt-2 border-t space-y-3'>
              <div className='flex items-center justify-between'>
                <div className='space-y-0.5'>
                  <Label className='text-xs font-medium'>Headline Banner</Label>
                  <p className='text-[10px] text-muted-foreground'>Feature on Homepage Hero</p>
                </div>
                <Switch
                  checked={isHeadline}
                  onCheckedChange={setIsHeadline}
                />
              </div>

              <div className='flex items-center justify-between'>
                <div className='space-y-0.5'>
                  <Label className='text-xs font-medium'>Breaking News</Label>
                  <p className='text-[10px] text-muted-foreground'>Show in Top Ticker Bar</p>
                </div>
                <Switch
                  checked={isBreaking}
                  onCheckedChange={setIsBreaking}
                />
              </div>
            </div>
          </div>

          {/* Taxonomies & Author */}
          <div className='p-4 border rounded-lg space-y-4 bg-card shadow-xs'>
            <h3 className='font-semibold text-sm border-b pb-2'>Taxonomy & Credits</h3>

            <div className='space-y-2'>
              <Label className='text-xs font-semibold'>
                Category <span className='text-red-500'>*</span>
              </Label>
              <Select
                value={categoryId ? String(categoryId) : undefined}
                onValueChange={(val) => setCategoryId(Number(val))}
              >
                <SelectTrigger className='w-full text-xs'>
                  <SelectValue placeholder='Select category' />
                </SelectTrigger>
                <SelectContent>
                  {categories.map((c) => (
                    <SelectItem key={c.id} value={String(c.id)}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-2'>
              <Label className='text-xs font-semibold'>
                Author Byline <span className='text-red-500'>*</span>
              </Label>
              <Select
                value={authorId ? String(authorId) : undefined}
                onValueChange={(val) => setAuthorId(Number(val))}
              >
                <SelectTrigger className='w-full text-xs'>
                  <SelectValue placeholder='Select author' />
                </SelectTrigger>
                <SelectContent>
                  {authors.map((a) => (
                    <SelectItem key={a.id} value={String(a.id)}>
                      {a.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-2 pt-2 border-t'>
              <Label className='text-xs font-semibold'>Tags</Label>
              <div className='flex flex-wrap gap-1.5 max-h-36 overflow-y-auto p-1 border rounded bg-muted/20'>
                {availableTags.map((tag) => {
                  const isSelected = selectedTagIds.includes(tag.id)
                  return (
                    <button
                      key={tag.id}
                      type='button'
                      onClick={() => handleToggleTag(tag.id)}
                      className={`text-[11px] px-2 py-0.5 rounded-full border transition-colors ${
                        isSelected
                          ? 'bg-primary text-primary-foreground border-primary font-medium'
                          : 'bg-background hover:bg-muted text-muted-foreground'
                      }`}
                    >
                      #{tag.name}
                    </button>
                  )
                })}
              </div>
            </div>
          </div>

          {/* Featured Cover Image */}
          <div className='p-4 border rounded-lg space-y-3 bg-card shadow-xs'>
            <h3 className='font-semibold text-sm border-b pb-2'>Featured Cover Image</h3>

            {coverMedia || initialArticle?.cover_url ? (
              <div className='space-y-2'>
                <div className='aspect-video rounded-lg overflow-hidden border bg-muted'>
                  <img
                    src={coverMedia?.variants?.medium || coverMedia?.file_path || initialArticle?.cover_url}
                    alt='Cover preview'
                    className='w-full h-full object-cover'
                  />
                </div>
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  className='w-full text-xs'
                  onClick={() => setIsMediaPickerOpen(true)}
                >
                  Change Cover Image
                </Button>
              </div>
            ) : (
              <div
                onClick={() => setIsMediaPickerOpen(true)}
                className='flex flex-col items-center justify-center p-6 border-2 border-dashed rounded-lg cursor-pointer hover:bg-muted/30 transition-colors gap-2 text-muted-foreground'
              >
                <ImageIcon className='h-8 w-8 stroke-[1.5]' />
                <span className='text-xs font-medium'>Select from Media Library</span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Media Picker Modal for Cover */}
      <MediaPickerDialog
        open={isMediaPickerOpen}
        onOpenChange={setIsMediaPickerOpen}
        onSelect={(media) => setCoverMedia(media)}
        title='Select Featured Cover Image'
      />

      {/* Media Picker Modal for TipTap inline insertion */}
      <MediaPickerDialog
        open={isEditorMediaPickerOpen}
        onOpenChange={setIsEditorMediaPickerOpen}
        onSelect={(media) => {
          const imgUrl = media.variants?.large || media.file_path
          setBody((prev) => `${prev}<p><img src="${imgUrl}" alt="${media.alt_text || media.filename}" /></p>`)
          toast.success('Image inserted into article body')
        }}
        title='Insert Image into Article Body'
      />
    </form>
  )
}
