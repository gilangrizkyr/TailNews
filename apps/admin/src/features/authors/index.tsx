import { useState, useEffect } from 'react'
import { Plus, Edit, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { authorApi } from '@/lib/api-client'
import { AuthorResponse, AuthorRequest } from '@/types/api'

export function AuthorsFeature() {
  const [authors, setAuthors] = useState<AuthorResponse[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [editingAuthor, setEditingAuthor] = useState<AuthorResponse | null>(null)

  // Form Fields
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [bio, setBio] = useState('')

  const loadAuthors = async () => {
    setIsLoading(true)
    try {
      const data = await authorApi.list()
      setAuthors(data)
    } catch {
      toast.error('Failed to load authors')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadAuthors()
  }, [])

  const handleOpenCreate = () => {
    setEditingAuthor(null)
    setName('')
    setSlug('')
    setBio('')
    setIsDialogOpen(true)
  }

  const handleOpenEdit = (author: AuthorResponse) => {
    setEditingAuthor(author)
    setName(author.name)
    setSlug(author.slug)
    setBio(author.bio || '')
    setIsDialogOpen(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      toast.error('Author name is required')
      return
    }

    setIsSubmitting(true)
    try {
      const payload: AuthorRequest = {
        name: name.trim(),
        slug: slug.trim() || undefined,
        bio: bio.trim() || undefined,
      }

      if (editingAuthor) {
        await authorApi.update(editingAuthor.id, payload)
        toast.success('Author profile updated')
      } else {
        await authorApi.create(payload)
        toast.success('Author profile created')
      }

      setIsDialogOpen(false)
      loadAuthors()
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || 'Failed to save author'
      toast.error(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className='p-6 space-y-6'>
      <div className='flex items-center justify-between'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>Authors & Journalists</h1>
          <p className='text-sm text-muted-foreground'>
            Manage editorial bylines, reporter bios, and author profiles ({authors.length} authors).
          </p>
        </div>
        <Button onClick={handleOpenCreate}>
          <Plus className='h-4 w-4 me-2' /> Add Author
        </Button>
      </div>

      <div className='border rounded-lg overflow-hidden bg-card'>
        <table className='w-full text-left text-xs'>
          <thead className='bg-muted/50 border-b text-muted-foreground uppercase tracking-wider font-semibold'>
            <tr>
              <th className='p-3'>Author Name & Slug</th>
              <th className='p-3'>Bio</th>
              <th className='p-3 w-40'>Created</th>
              <th className='p-3 w-20 text-center'>Actions</th>
            </tr>
          </thead>
          <tbody className='divide-y divide-border'>
            {isLoading ? (
              <tr>
                <td colSpan={4} className='p-12 text-center text-muted-foreground'>
                  <Loader2 className='h-6 w-6 animate-spin mx-auto mb-2' />
                  Loading authors...
                </td>
              </tr>
            ) : authors.length === 0 ? (
              <tr>
                <td colSpan={4} className='p-12 text-center text-muted-foreground'>
                  No authors registered.
                </td>
              </tr>
            ) : (
              authors.map((a) => (
                <tr key={a.id} className='hover:bg-muted/30 transition-colors'>
                  <td className='p-3'>
                    <div className='font-semibold text-sm'>{a.name}</div>
                    <code className='text-[11px] text-muted-foreground font-mono'>
                      /author/{a.slug}
                    </code>
                  </td>
                  <td className='p-3 text-muted-foreground max-w-md line-clamp-1'>
                    {a.bio || '—'}
                  </td>
                  <td className='p-3 text-muted-foreground'>
                    {new Date(a.created_at).toLocaleDateString()}
                  </td>
                  <td className='p-3 text-center'>
                    <Button
                      size='sm'
                      variant='ghost'
                      className='h-7 w-7 p-0'
                      onClick={() => handleOpenEdit(a)}
                    >
                      <Edit className='h-3.5 w-3.5' />
                    </Button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Author Modal */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle>
              {editingAuthor ? 'Edit Author Profile' : 'New Author Profile'}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className='space-y-4 py-2 text-xs'>
            <div className='space-y-1.5'>
              <Label htmlFor='authName'>Full Name</Label>
              <Input
                id='authName'
                placeholder='e.g. Rina Wijaya'
                value={name}
                onChange={(e) => {
                  setName(e.target.value)
                  if (!editingAuthor) {
                    setSlug(
                      e.target.value
                        .toLowerCase()
                        .replace(/[^\w\s-]/g, '')
                        .trim()
                        .replace(/\s+/g, '-')
                    )
                  }
                }}
                required
              />
            </div>

            <div className='space-y-1.5'>
              <Label htmlFor='authSlug'>Slug</Label>
              <Input
                id='authSlug'
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
              />
            </div>

            <div className='space-y-1.5'>
              <Label htmlFor='authBio'>Biography</Label>
              <Textarea
                id='authBio'
                placeholder='Journalist bio, beats, credentials...'
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                className='min-h-[80px]'
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
                {editingAuthor ? 'Save Changes' : 'Create Profile'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
