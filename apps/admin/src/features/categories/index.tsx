import { useState, useEffect } from 'react'
import {
  Plus,
  Edit,
  Trash2,
  Loader2,
  Check,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { categoryApi } from '@/lib/api-client'
import { CategoryResponse, CategoryRequest } from '@/types/api'

export function CategoriesFeature() {
  const [categories, setCategories] = useState<CategoryResponse[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [editingCategory, setEditingCategory] = useState<CategoryResponse | null>(null)

  // Form Fields
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [description, setDescription] = useState('')
  const [isNav, setIsNav] = useState(true)
  const [sortOrder, setSortOrder] = useState(0)

  const loadCategories = async () => {
    setIsLoading(true)
    try {
      const data = await categoryApi.list()
      setCategories(data)
    } catch {
      toast.error('Failed to load categories')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadCategories()
  }, [])

  const handleOpenCreate = () => {
    setEditingCategory(null)
    setName('')
    setSlug('')
    setDescription('')
    setIsNav(true)
    setSortOrder(0)
    setIsDialogOpen(true)
  }

  const handleOpenEdit = (cat: CategoryResponse) => {
    setEditingCategory(cat)
    setName(cat.name)
    setSlug(cat.slug)
    setDescription(cat.description || '')
    setIsNav(cat.is_nav)
    setSortOrder(cat.sort_order)
    setIsDialogOpen(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      toast.error('Category name is required')
      return
    }

    setIsSubmitting(true)
    try {
      const payload: CategoryRequest = {
        name: name.trim(),
        slug: slug.trim() || undefined,
        description: description.trim() || undefined,
        is_nav: isNav,
        sort_order: Number(sortOrder),
      }

      if (editingCategory) {
        await categoryApi.update(editingCategory.id, payload)
        toast.success('Category updated successfully')
      } else {
        await categoryApi.create(payload)
        toast.success('Category created successfully')
      }

      setIsDialogOpen(false)
      loadCategories()
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || 'Failed to save category'
      toast.error(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this category?')) return
    try {
      await categoryApi.delete(id)
      toast.success('Category deleted')
      loadCategories()
    } catch {
      toast.error('Failed to delete category')
    }
  }

  return (
    <div className='p-6 space-y-6'>
      <div className='flex items-center justify-between'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>Categories</h1>
          <p className='text-sm text-muted-foreground'>
            Organize articles into dynamic sections, topics, and navigation channels.
          </p>
        </div>
        <Button onClick={handleOpenCreate}>
          <Plus className='h-4 w-4 me-2' /> Add Category
        </Button>
      </div>

      <div className='border rounded-lg overflow-hidden bg-card'>
        <table className='w-full text-left text-xs'>
          <thead className='bg-muted/50 border-b text-muted-foreground uppercase tracking-wider font-semibold'>
            <tr>
              <th className='p-3'>Name & Slug</th>
              <th className='p-3'>Description</th>
              <th className='p-3 w-28 text-center'>Navbar Item</th>
              <th className='p-3 w-24 text-center'>Sort Order</th>
              <th className='p-3 w-24 text-center'>Actions</th>
            </tr>
          </thead>
          <tbody className='divide-y divide-border'>
            {isLoading ? (
              <tr>
                <td colSpan={5} className='p-12 text-center text-muted-foreground'>
                  <Loader2 className='h-6 w-6 animate-spin mx-auto mb-2' />
                  Loading categories...
                </td>
              </tr>
            ) : categories.length === 0 ? (
              <tr>
                <td colSpan={5} className='p-12 text-center text-muted-foreground'>
                  No categories found.
                </td>
              </tr>
            ) : (
              categories.map((cat) => (
                <tr key={cat.id} className='hover:bg-muted/30 transition-colors'>
                  <td className='p-3'>
                    <div className='font-semibold text-sm'>{cat.name}</div>
                    <code className='text-[11px] text-muted-foreground font-mono'>
                      /{cat.slug}
                    </code>
                  </td>
                  <td className='p-3 text-muted-foreground max-w-sm truncate'>
                    {cat.description || '—'}
                  </td>
                  <td className='p-3 text-center'>
                    {cat.is_nav ? (
                      <Badge variant='outline' className='bg-green-500/10 text-green-600 border-green-500/20'>
                        <Check className='h-3 w-3 me-1' /> Yes
                      </Badge>
                    ) : (
                      <Badge variant='outline' className='text-muted-foreground'>
                        <X className='h-3 w-3 me-1' /> No
                      </Badge>
                    )}
                  </td>
                  <td className='p-3 text-center font-mono font-medium'>
                    {cat.sort_order}
                  </td>
                  <td className='p-3 text-center'>
                    <div className='flex items-center justify-center gap-1'>
                      <Button
                        size='sm'
                        variant='ghost'
                        className='h-7 w-7 p-0'
                        onClick={() => handleOpenEdit(cat)}
                      >
                        <Edit className='h-3.5 w-3.5' />
                      </Button>
                      <Button
                        size='sm'
                        variant='ghost'
                        className='h-7 w-7 p-0 text-destructive'
                        onClick={() => handleDelete(cat.id)}
                      >
                        <Trash2 className='h-3.5 w-3.5' />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Category Modal */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle>
              {editingCategory ? 'Edit Category' : 'Create Category'}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className='space-y-4 py-2 text-xs'>
            <div className='space-y-1.5'>
              <Label htmlFor='catName'>Category Name</Label>
              <Input
                id='catName'
                placeholder='e.g. Ekonomi, Teknologi, Politik'
                value={name}
                onChange={(e) => {
                  setName(e.target.value)
                  if (!editingCategory) {
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
              <Label htmlFor='catSlug'>Slug</Label>
              <Input
                id='catSlug'
                placeholder='url-slug'
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
              />
            </div>

            <div className='space-y-1.5'>
              <Label htmlFor='catDesc'>Description</Label>
              <Textarea
                id='catDesc'
                placeholder='Brief description of this editorial section...'
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>

            <div className='flex items-center justify-between pt-2'>
              <div>
                <Label className='text-xs'>Show in Navigation Bar</Label>
                <p className='text-[10px] text-muted-foreground'>
                  Include in the top header menu of the public site.
                </p>
              </div>
              <Switch checked={isNav} onCheckedChange={setIsNav} />
            </div>

            <div className='space-y-1.5'>
              <Label htmlFor='catOrder'>Sort Order</Label>
              <Input
                id='catOrder'
                type='number'
                value={sortOrder}
                onChange={(e) => setSortOrder(Number(e.target.value))}
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
                {editingCategory ? 'Save Changes' : 'Create Category'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
