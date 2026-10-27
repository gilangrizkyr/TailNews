import { useState, useEffect } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { History, RotateCcw, Loader2, Calendar } from 'lucide-react'
import { articleApi } from '@/lib/api-client'
import { ArticleRevisionResponse } from '@/types/api'

interface RevisionsDialogProps {
  articleId: number
  articleTitle: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onRestored: () => void
}

export function RevisionsDialog({
  articleId,
  articleTitle,
  open,
  onOpenChange,
  onRestored,
}: RevisionsDialogProps) {
  const [revisions, setRevisions] = useState<ArticleRevisionResponse[]>([])
  const [selectedRevision, setSelectedRevision] = useState<ArticleRevisionResponse | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isRestoring, setIsRestoring] = useState(false)

  const loadRevisions = async () => {
    setIsLoading(true)
    try {
      const data = await articleApi.listRevisions(articleId)
      setRevisions(data)
      if (data.length > 0) {
        setSelectedRevision(data[0])
      }
    } catch {
      toast.error('Failed to load article revisions')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (open) {
      loadRevisions()
    }
  }, [open, articleId])

  const handleRestore = async (revId: number) => {
    if (!confirm('Are you sure you want to restore this revision? It will create a new current version.')) return

    setIsRestoring(true)
    try {
      await articleApi.restoreRevision(articleId, revId)
      toast.success('Article revision restored successfully!')
      onRestored()
      onOpenChange(false)
    } catch {
      toast.error('Failed to restore revision')
    } finally {
      setIsRestoring(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-4xl max-h-[85vh] flex flex-col p-6'>
        <DialogHeader className='pb-3 border-b'>
          <DialogTitle className='flex items-center gap-2 text-lg'>
            <History className='h-5 w-5 text-primary' />
            Revision History: <span className='font-normal truncate'>{articleTitle}</span>
          </DialogTitle>
        </DialogHeader>

        {isLoading ? (
          <div className='flex flex-col items-center justify-center h-64 gap-2 text-muted-foreground'>
            <Loader2 className='h-8 w-8 animate-spin' />
            <p>Loading historical revisions...</p>
          </div>
        ) : revisions.length === 0 ? (
          <div className='flex flex-col items-center justify-center h-48 text-muted-foreground gap-2'>
            <History className='h-10 w-10 stroke-[1.5]' />
            <p>No prior revisions recorded for this article yet.</p>
          </div>
        ) : (
          <div className='grid grid-cols-1 md:grid-cols-3 gap-4 flex-1 overflow-hidden py-3'>
            {/* List of Revisions */}
            <div className='border rounded-lg overflow-y-auto max-h-[500px] divide-y'>
              {revisions.map((rev) => {
                const isSelected = selectedRevision?.id === rev.id
                return (
                  <div
                    key={rev.id}
                    onClick={() => setSelectedRevision(rev)}
                    className={`p-3 cursor-pointer transition-colors ${
                      isSelected
                        ? 'bg-primary/10 border-s-4 border-s-primary'
                        : 'hover:bg-muted/50'
                    }`}
                  >
                    <div className='flex items-center justify-between text-xs text-muted-foreground mb-1'>
                      <span className='flex items-center gap-1'>
                        <Calendar className='h-3 w-3' />
                        {new Date(rev.created_at).toLocaleString()}
                      </span>
                    </div>
                    <p className='font-medium text-sm line-clamp-1'>{rev.title}</p>
                    <p className='text-xs text-muted-foreground line-clamp-2 mt-0.5'>
                      {rev.excerpt}
                    </p>
                  </div>
                )
              })}
            </div>

            {/* Revision Preview */}
            <div className='md:col-span-2 border rounded-lg p-4 flex flex-col overflow-y-auto max-h-[500px] space-y-4 bg-muted/20'>
              {selectedRevision ? (
                <>
                  <div className='flex items-center justify-between pb-3 border-b'>
                    <div>
                      <h3 className='font-semibold text-base'>{selectedRevision.title}</h3>
                      <p className='text-xs text-muted-foreground'>
                        Slug: <code className='text-xs font-mono'>{selectedRevision.slug}</code>
                      </p>
                    </div>
                    <Button
                      size='sm'
                      disabled={isRestoring}
                      onClick={() => handleRestore(selectedRevision.id)}
                    >
                      {isRestoring ? (
                        <Loader2 className='h-4 w-4 animate-spin me-1.5' />
                      ) : (
                        <RotateCcw className='h-4 w-4 me-1.5' />
                      )}
                      Restore This Version
                    </Button>
                  </div>

                  <div className='space-y-2'>
                    <span className='text-xs font-bold text-muted-foreground uppercase tracking-wide'>
                      Excerpt Summary:
                    </span>
                    <p className='text-sm italic bg-background p-3 rounded border'>
                      {selectedRevision.excerpt}
                    </p>
                  </div>

                  <div className='space-y-2'>
                    <span className='text-xs font-bold text-muted-foreground uppercase tracking-wide'>
                      HTML Content Snapshot:
                    </span>
                    <div
                      className='prose dark:prose-invert max-w-none text-sm p-3 bg-background rounded border overflow-y-auto max-h-64'
                      dangerouslySetInnerHTML={{ __html: selectedRevision.body }}
                    />
                  </div>
                </>
              ) : (
                <div className='flex items-center justify-center h-full text-muted-foreground text-sm'>
                  Select a revision from the left to preview.
                </div>
              )}
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
