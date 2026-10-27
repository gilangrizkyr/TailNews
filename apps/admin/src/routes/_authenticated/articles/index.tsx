import { createFileRoute } from '@tanstack/react-router'
import { ArticlesFeature } from '@/features/articles'

export const Route = createFileRoute('/_authenticated/articles/')({
  component: ArticlesFeature,
})
