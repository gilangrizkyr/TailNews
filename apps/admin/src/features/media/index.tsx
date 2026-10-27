import { useState, useEffect, useRef } from 'react'
import {
  UploadCloud,
  Trash2,
  Copy,
  Check,
  Loader2,
  Image as ImageIcon,
  Search,
} from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { mediaApi } from '@/lib/api-client'
import { MediaResponse } from '@/types/api'

export function MediaLibraryFeature() {
  const [mediaList, setMediaList] = useState<MediaResponse[]>([])
  const [selectedMedia, setSelectedMedia] = useState<MediaResponse | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  const [search, setSearch] = useState('')
  const [copiedUrl, setCopiedUrl] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const fetchMedia = async () => {
    setIsLoading(true)
    try {
      const res = await mediaApi.list(1, 50)
      setMediaList(res.data)
    } catch {
      toast.error('Failed to load media assets')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchMedia()
  }, [])

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setIsUploading(true)
    try {
      const uploaded = await mediaApi.upload(file, file.name, file.name)
      toast.success('Image successfully uploaded and variants processed!')
      setMediaList((prev) => [uploaded, ...prev])
      setSelectedMedia(uploaded)
    } catch {
      toast.error('Upload failed. Only JPG, PNG, and WebP are supported.')
    } finally {
      setIsUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this media asset?')) return
    try {
      await mediaApi.delete(id)
      toast.success('Media asset deleted')
      setMediaList((prev) => prev.filter((m) => m.id !== id))
      if (selectedMedia?.id === id) setSelectedMedia(null)
    } catch {
      toast.error('Failed to delete media asset')
    }
  }

  const copyToClipboard = (url: string) => {
    navigator.clipboard.writeText(url)
    setCopiedUrl(url)
    toast.success('Image URL copied to clipboard')
    setTimeout(() => setCopiedUrl(null), 2000)
  }

  const filteredMedia = mediaList.filter((m) =>
    m.filename.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className='p-6 space-y-6'>
      <div className='flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>Media Library</h1>
          <p className='text-muted-foreground text-sm'>
            Upload, optimize, and organize high-resolution images with automatic WebP variants.
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <input
            type='file'
            ref={fileInputRef}
            onChange={handleUpload}
            accept='image/jpeg,image/png,image/webp'
            className='hidden'
          />
          <Button
            disabled={isUploading}
            onClick={() => fileInputRef.current?.click()}
          >
            {isUploading ? (
              <Loader2 className='h-4 w-4 animate-spin me-2' />
            ) : (
              <UploadCloud className='h-4 w-4 me-2' />
            )}
            Upload Image
          </Button>
        </div>
      </div>

      <div className='flex items-center gap-3'>
        <div className='relative flex-1 max-w-sm'>
          <Search className='absolute left-3 top-2.5 h-4 w-4 text-muted-foreground' />
          <Input
            placeholder='Search media by filename...'
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className='pl-9'
          />
        </div>
      </div>

      {isLoading ? (
        <div className='flex flex-col items-center justify-center h-64 gap-3 text-muted-foreground'>
          <Loader2 className='h-8 w-8 animate-spin' />
          <p>Loading media assets...</p>
        </div>
      ) : filteredMedia.length === 0 ? (
        <div className='flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-xl gap-3 text-muted-foreground'>
          <ImageIcon className='h-12 w-12 stroke-[1.5]' />
          <p className='text-base font-medium'>No media assets found</p>
          <Button
            variant='outline'
            onClick={() => fileInputRef.current?.click()}
          >
            Upload your first photo
          </Button>
        </div>
      ) : (
        <div className='grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4'>
          {filteredMedia.map((item) => {
            const previewUrl =
              item.variants?.thumbnail || item.file_path || ''
            return (
              <div
                key={item.id}
                onClick={() => setSelectedMedia(item)}
                className='group relative aspect-square rounded-lg border overflow-hidden bg-muted cursor-pointer hover:shadow-md transition-all'
              >
                <img
                  src={previewUrl}
                  alt={item.alt_text || item.filename}
                  className='w-full h-full object-cover transition-transform duration-200 group-hover:scale-105'
                  loading='lazy'
                />
                <div className='absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity p-2.5 flex flex-col justify-end'>
                  <p className='text-xs font-semibold text-white truncate'>
                    {item.filename}
                  </p>
                  <p className='text-[10px] text-white/80'>
                    {item.width} x {item.height} • {Math.round(item.size_bytes / 1024)} KB
                  </p>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Selected Media Detail Modal */}
      {selectedMedia && (
        <Dialog
          open={!!selectedMedia}
          onOpenChange={(open) => !open && setSelectedMedia(null)}
        >
          <DialogContent className='sm:max-w-2xl'>
            <DialogHeader>
              <DialogTitle className='truncate text-base'>
                {selectedMedia.filename}
              </DialogTitle>
            </DialogHeader>

            <div className='grid grid-cols-1 sm:grid-cols-2 gap-4 py-2'>
              <div className='aspect-video rounded-lg overflow-hidden border bg-black/5 flex items-center justify-center'>
                <img
                  src={selectedMedia.variants?.medium || selectedMedia.file_path}
                  alt={selectedMedia.alt_text}
                  className='max-h-full object-contain'
                />
              </div>

              <div className='space-y-3 text-xs'>
                <div>
                  <span className='font-semibold text-muted-foreground'>Dimensions:</span>
                  <p className='font-mono'>{selectedMedia.width} x {selectedMedia.height} px</p>
                </div>
                <div>
                  <span className='font-semibold text-muted-foreground'>File Size:</span>
                  <p className='font-mono'>{Math.round(selectedMedia.size_bytes / 1024)} KB</p>
                </div>
                <div>
                  <span className='font-semibold text-muted-foreground'>MIME Type:</span>
                  <p className='font-mono uppercase'>{selectedMedia.mime_type}</p>
                </div>

                <div className='space-y-1.5 pt-2 border-t'>
                  <span className='font-semibold text-muted-foreground'>Responsive Variants:</span>
                  <div className='flex flex-wrap gap-1'>
                    {Object.entries(selectedMedia.variants || {}).map(([name, url]) => (
                      <Button
                        key={name}
                        size='sm'
                        variant='outline'
                        className='text-[10px] h-7 px-2'
                        onClick={() => copyToClipboard(url)}
                      >
                        {copiedUrl === url ? (
                          <Check className='h-3 w-3 me-1 text-green-500' />
                        ) : (
                          <Copy className='h-3 w-3 me-1' />
                        )}
                        {name}
                      </Button>
                    ))}
                  </div>
                </div>

                <div className='flex items-center gap-2 pt-3'>
                  <Button
                    size='sm'
                    variant='outline'
                    className='w-full'
                    onClick={() => copyToClipboard(selectedMedia.file_path)}
                  >
                    {copiedUrl === selectedMedia.file_path ? (
                      <Check className='h-3.5 w-3.5 me-1.5 text-green-500' />
                    ) : (
                      <Copy className='h-3.5 w-3.5 me-1.5' />
                    )}
                    Copy Original URL
                  </Button>
                  <Button
                    size='sm'
                    variant='destructive'
                    onClick={() => handleDelete(selectedMedia.id)}
                  >
                    <Trash2 className='h-3.5 w-3.5' />
                  </Button>
                </div>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}
