import { createFileRoute } from '@tanstack/react-router'
import { MediaLibraryFeature } from '@/features/media'

export const Route = createFileRoute('/_authenticated/media/')({
  component: MediaLibraryFeature,
})
