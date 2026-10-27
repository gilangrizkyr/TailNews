import { useState, useEffect } from 'react'
import { Plus, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { redirectApi } from '@/lib/api-client'
import { RedirectResponse, RedirectRequest } from '@/types/api'

export function RedirectsFeature() {
  const [redirects, setRedirects] = useState<RedirectResponse[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Form Fields
  const [fromPath, setFromPath] = useState('')
  const [toPath, setToPath] = useState('')
  const [httpStatus, setHttpStatus] = useState<number>(301)

  const loadRedirects = async () => {
    setIsLoading(true)
    try {
      const data = await redirectApi.list()
      setRedirects(data)
    } catch {
      toast.error('Failed to load redirects')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadRedirects()
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!fromPath.startsWith('/') || !toPath.startsWith('/')) {
      toast.error('Both from and to paths must start with a forward slash (/), e.g. /old-slug')
      return
    }

    setIsSubmitting(true)
    try {
      const payload: RedirectRequest = {
        from_path: fromPath.trim(),
        to_path: toPath.trim(),
        http_status: httpStatus,
      }
      await redirectApi.create(payload)
      toast.success('Redirect rule registered successfully')
      setIsDialogOpen(false)
      setFromPath('')
      setToPath('')
      loadRedirects()
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || 'Failed to register redirect'
      toast.error(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className='p-6 space-y-6'>
      <div className='flex items-center justify-between'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>301 & 302 Redirect Rules</h1>
          <p className='text-sm text-muted-foreground'>
            Prevent 404 dead-ends and preserve SEO link equity across slug updates ({redirects.length} rules).
          </p>
        </div>
        <Button onClick={() => setIsDialogOpen(true)}>
          <Plus className='h-4 w-4 me-2' /> Add Redirect Rule
        </Button>
      </div>

      <div className='border rounded-lg overflow-hidden bg-card'>
        <table className='w-full text-left text-xs'>
          <thead className='bg-muted/50 border-b text-muted-foreground uppercase tracking-wider font-semibold'>
            <tr>
              <th className='p-3'>Source Path (From)</th>
              <th className='p-3'>Target Path (To)</th>
              <th className='p-3 w-28 text-center'>HTTP Status</th>
              <th className='p-3 w-28 text-right'>Hits</th>
              <th className='p-3 w-36'>Created</th>
            </tr>
          </thead>
          <tbody className='divide-y divide-border font-mono'>
            {isLoading ? (
              <tr>
                <td colSpan={5} className='p-12 text-center text-muted-foreground font-sans'>
                  <Loader2 className='h-6 w-6 animate-spin mx-auto mb-2' />
                  Loading redirect rules...
                </td>
              </tr>
            ) : redirects.length === 0 ? (
              <tr>
                <td colSpan={5} className='p-12 text-center text-muted-foreground font-sans'>
                  No redirects registered yet.
                </td>
              </tr>
            ) : (
              redirects.map((r) => (
                <tr key={r.id} className='hover:bg-muted/30 transition-colors'>
                  <td className='p-3 text-red-600 dark:text-red-400 font-semibold truncate'>
                    {r.from_path}
                  </td>
                  <td className='p-3 text-green-600 dark:text-green-400 font-semibold truncate'>
                    {r.to_path}
                  </td>
                  <td className='p-3 text-center'>
                    <Badge variant='outline' className='text-xs font-mono'>
                      {r.http_status} Permanent
                    </Badge>
                  </td>
                  <td className='p-3 text-right font-medium text-muted-foreground'>
                    {r.hit_count.toLocaleString()}
                  </td>
                  <td className='p-3 text-muted-foreground font-sans'>
                    {new Date(r.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Add Modal */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle>Add URL Redirect Rule</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className='space-y-4 py-2 text-xs'>
            <div className='space-y-1.5'>
              <Label htmlFor='fromPath'>From Path (Old URL)</Label>
              <Input
                id='fromPath'
                placeholder='/berita/old-article-slug'
                value={fromPath}
                onChange={(e) => setFromPath(e.target.value)}
                required
              />
            </div>

            <div className='space-y-1.5'>
              <Label htmlFor='toPath'>To Path (New Destination URL)</Label>
              <Input
                id='toPath'
                placeholder='/berita/new-article-slug'
                value={toPath}
                onChange={(e) => setToPath(e.target.value)}
                required
              />
            </div>

            <div className='space-y-1.5'>
              <Label>HTTP Status Code</Label>
              <Select
                value={String(httpStatus)}
                onValueChange={(val) => setHttpStatus(Number(val))}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='301'>301 Moved Permanently (Recommended for SEO)</SelectItem>
                  <SelectItem value='302'>302 Found / Temporary</SelectItem>
                </SelectContent>
              </Select>
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
                Create Redirect
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
