import { useState, useEffect } from 'react'
import { Tag as TagIcon, Plus, Trash2, Loader2, Search } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { tagApi } from '@/lib/api-client'
import { TagResponse } from '@/types/api'

export function TagsFeature() {
  const [tags, setTags] = useState<TagResponse[]>([])
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')

  const loadTags = async () => {
    setIsLoading(true)
    try {
      const data = await tagApi.list()
      setTags(data)
    } catch {
      toast.error('Failed to load tags')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadTags()
  }, [])

  const handleOpenCreate = () => {
    setName('')
    setSlug('')
    setIsDialogOpen(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      toast.error('Tag name is required')
      return
    }

    setIsSubmitting(true)
    try {
      await tagApi.create({
        name: name.trim(),
        slug: slug.trim() || undefined,
      })
      toast.success('Tag created successfully')
      setIsDialogOpen(false)
      loadTags()
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || 'Failed to create tag'
      toast.error(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this tag?')) return
    try {
      await tagApi.delete(id)
      toast.success('Tag deleted')
      setTags((prev) => prev.filter((t) => t.id !== id))
    } catch {
      toast.error('Failed to delete tag')
    }
  }

  const filteredTags = tags.filter((t) =>
    t.name.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className='p-6 space-y-6'>
      <div className='flex items-center justify-between'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>Article Tags</h1>
          <p className='text-sm text-muted-foreground'>
            Topic hashtags and keywords for deep categorization and search indexing ({tags.length} total).
          </p>
        </div>
        <Button onClick={handleOpenCreate}>
          <Plus className='h-4 w-4 me-2' /> Add Tag
        </Button>
      </div>

      <div className='relative max-w-sm'>
        <Search className='absolute left-3 top-2.5 h-4 w-4 text-muted-foreground' />
        <Input
          placeholder='Filter tags...'
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className='pl-9 h-9'
        />
      </div>

      <div className='border rounded-lg overflow-hidden bg-card'>
        <table className='w-full text-left text-xs'>
          <thead className='bg-muted/50 border-b text-muted-foreground uppercase tracking-wider font-semibold'>
            <tr>
              <th className='p-3'>Tag Name</th>
              <th className='p-3'>Slug</th>
              <th className='p-3 w-40'>Created</th>
              <th className='p-3 w-20 text-center'>Actions</th>
            </tr>
          </thead>
          <tbody className='divide-y divide-border'>
            {isLoading ? (
              <tr>
                <td colSpan={4} className='p-12 text-center text-muted-foreground'>
                  <Loader2 className='h-6 w-6 animate-spin mx-auto mb-2' />
                  Loading tags...
                </td>
              </tr>
            ) : filteredTags.length === 0 ? (
              <tr>
                <td colSpan={4} className='p-12 text-center text-muted-foreground'>
                  No tags found.
                </td>
              </tr>
            ) : (
              filteredTags.slice(0, 100).map((t) => (
                <tr key={t.id} className='hover:bg-muted/30 transition-colors'>
                  <td className='p-3 font-semibold flex items-center gap-1.5'>
                    <TagIcon className='h-3.5 w-3.5 text-muted-foreground' />
                    #{t.name}
                  </td>
                  <td className='p-3 font-mono text-muted-foreground'>/{t.slug}</td>
                  <td className='p-3 text-muted-foreground'>
                    {new Date(t.created_at).toLocaleDateString()}
                  </td>
                  <td className='p-3 text-center'>
                    <Button
                      size='sm'
                      variant='ghost'
                      className='h-7 w-7 p-0 text-destructive'
                      onClick={() => handleDelete(t.id)}
                    >
                      <Trash2 className='h-3.5 w-3.5' />
                    </Button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Create Modal */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle>Create New Tag</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className='space-y-4 py-2 text-xs'>
            <div className='space-y-1.5'>
              <Label htmlFor='tagName'>Tag Name</Label>
              <Input
                id='tagName'
                placeholder='e.g. artificial-intelligence, iis-indonesia'
                value={name}
                onChange={(e) => {
                  setName(e.target.value)
                  setSlug(
                    e.target.value
                      .toLowerCase()
                      .replace(/[^\w\s-]/g, '')
                      .trim()
                      .replace(/\s+/g, '-')
                  )
                }}
                required
              />
            </div>
            <div className='space-y-1.5'>
              <Label htmlFor='tagSlug'>Slug</Label>
              <Input
                id='tagSlug'
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
              />
            </div>

            <DialogFooter className='pt-3 border-t'>
              <Button
                type='button'
                variant='outline'
                onClick={() => setIsDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button type='submit' disabled={isSubmitting}>
                {isSubmitting && <Loader2 className='h-4 w-4 animate-spin me-1.5' />}
                Create Tag
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
