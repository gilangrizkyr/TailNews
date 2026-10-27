import { createFileRoute } from '@tanstack/react-router'
import { ArticleForm } from '@/features/articles/components/article-form'

export const Route = createFileRoute('/_authenticated/articles/new')({
  component: NewArticlePage,
})

function NewArticlePage() {
  return <ArticleForm isEdit={false} />
}
