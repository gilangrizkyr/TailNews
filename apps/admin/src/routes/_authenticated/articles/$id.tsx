import { useState, useEffect } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { ArticleForm } from '@/features/articles/components/article-form'
import { articleApi } from '@/lib/api-client'
import { ArticleResponse } from '@/types/api'

export const Route = createFileRoute('/_authenticated/articles/$id')({
  component: EditArticlePage,
})

function EditArticlePage() {
  const { id } = Route.useParams()
  const [article, setArticle] = useState<ArticleResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const fetchArticle = async () => {
      try {
        const data = await articleApi.get(Number(id))
        setArticle(data)
      } catch {
        toast.error('Failed to load article details')
      } finally {
        setIsLoading(false)
      }
    }
    fetchArticle()
  }, [id])

  if (isLoading) {
    return (
      <div className='flex flex-col items-center justify-center h-96 gap-3 text-muted-foreground'>
        <Loader2 className='h-8 w-8 animate-spin' />
        <p className='text-sm'>Loading article #{id}...</p>
      </div>
    )
  }

  if (!article) {
    return (
      <div className='p-8 text-center text-muted-foreground'>
        <p>Article not found.</p>
      </div>
    )
  }

  return <ArticleForm initialArticle={article} isEdit={true} />
}
