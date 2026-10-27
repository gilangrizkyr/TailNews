import { createFileRoute } from '@tanstack/react-router'
import { TagsFeature } from '@/features/tags'

export const Route = createFileRoute('/_authenticated/tags/')({
  component: TagsFeature,
})
