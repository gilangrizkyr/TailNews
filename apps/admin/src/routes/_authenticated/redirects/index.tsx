import { createFileRoute } from '@tanstack/react-router'
import { RedirectsFeature } from '@/features/redirects'

export const Route = createFileRoute('/_authenticated/redirects/')({
  component: RedirectsFeature,
})
