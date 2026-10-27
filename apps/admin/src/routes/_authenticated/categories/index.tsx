import { createFileRoute } from '@tanstack/react-router'
import { CategoriesFeature } from '@/features/categories'

export const Route = createFileRoute('/_authenticated/categories/')({
  component: CategoriesFeature,
})
