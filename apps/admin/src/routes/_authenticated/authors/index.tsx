import { createFileRoute } from '@tanstack/react-router'
import { AuthorsFeature } from '@/features/authors'

export const Route = createFileRoute('/_authenticated/authors/')({
  component: AuthorsFeature,
})
