import { useState, useEffect, useRef } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { UploadCloud, Check, Loader2, Image as ImageIcon } from 'lucide-react'
import { mediaApi } from '@/lib/api-client'
import { MediaResponse } from '@/types/api'

interface MediaPickerDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSelect: (media: MediaResponse) => void
  title?: string
}

export function MediaPickerDialog({
  open,
  onOpenChange,
  onSelect,
  title = 'Select Media Asset',
}: MediaPickerDialogProps) {
  const [mediaList, setMediaList] = useState<MediaResponse[]>([])
  const [selectedMedia, setSelectedMedia] = useState<MediaResponse | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const loadMedia = async () => {
    setIsLoading(true)
    try {
      const res = await mediaApi.list(1, 40)
      setMediaList(res.data)
    } catch {
      toast.error('Failed to load media assets')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (open) {
      loadMedia()
    }
  }, [open])

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setIsUploading(true)
    try {
      const uploaded = await mediaApi.upload(file, file.name, file.name)
      toast.success('Media uploaded successfully!')
      setMediaList((prev) => [uploaded, ...prev])
      setSelectedMedia(uploaded)
    } catch {
      toast.error('Failed to upload image. Must be JPEG, PNG, or WebP.')
    } finally {
      setIsUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  const handleConfirm = () => {
    if (selectedMedia) {
      onSelect(selectedMedia)
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-4xl max-h-[85vh] flex flex-col p-6'>
        <DialogHeader className='flex flex-row items-center justify-between pb-2 border-b'>
          <DialogTitle className='text-lg font-bold'>{title}</DialogTitle>
          <div className='flex items-center gap-2'>
            <input
              type='file'
              ref={fileInputRef}
              onChange={handleFileUpload}
              accept='image/jpeg,image/png,image/webp'
              className='hidden'
            />
            <Button
              size='sm'
              disabled={isUploading}
              onClick={() => fileInputRef.current?.click()}
            >
              {isUploading ? (
                <Loader2 className='h-4 w-4 animate-spin me-2' />
              ) : (
                <UploadCloud className='h-4 w-4 me-2' />
              )}
              Upload New
            </Button>
          </div>
        </DialogHeader>

        <div className='flex-1 overflow-y-auto py-4'>
          {isLoading ? (
            <div className='flex flex-col items-center justify-center h-64 gap-2 text-muted-foreground'>
              <Loader2 className='h-8 w-8 animate-spin' />
              <span>Loading media library...</span>
            </div>
          ) : mediaList.length === 0 ? (
            <div className='flex flex-col items-center justify-center h-64 gap-3 text-muted-foreground border-2 border-dashed rounded-lg'>
              <ImageIcon className='h-12 w-12 stroke-[1.5]' />
              <p className='text-sm'>No media uploaded yet.</p>
              <Button
                variant='outline'
                size='sm'
                onClick={() => fileInputRef.current?.click()}
              >
                Upload First Photo
              </Button>
            </div>
          ) : (
            <div className='grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4'>
              {mediaList.map((item) => {
                const isSelected = selectedMedia?.id === item.id
                const previewUrl =
                  item.variants?.thumbnail || item.file_path || ''
                return (
                  <div
                    key={item.id}
                    onClick={() => setSelectedMedia(item)}
                    className={`group relative rounded-lg border overflow-hidden cursor-pointer aspect-video bg-muted flex items-center justify-center transition-all ${
                      isSelected
                        ? 'ring-2 ring-primary border-primary shadow-md'
                        : 'hover:border-foreground/40'
                    }`}
                  >
                    <img
                      src={previewUrl}
                      alt={item.alt_text || item.filename}
                      className='w-full h-full object-cover'
                      loading='lazy'
                    />
                    {isSelected && (
                      <div className='absolute top-2 right-2 bg-primary text-primary-foreground p-1 rounded-full shadow'>
                        <Check className='h-3.5 w-3.5 stroke-[3]' />
                      </div>
                    )}
                    <div className='absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 to-transparent p-2 opacity-0 group-hover:opacity-100 transition-opacity'>
                      <p className='text-xs text-white truncate font-medium'>
                        {item.filename}
                      </p>
                      <p className='text-[10px] text-white/70'>
                        {item.width}x{item.height} px
                      </p>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>

        {selectedMedia && (
          <div className='p-3 bg-muted/40 rounded-lg flex items-center justify-between text-xs'>
            <div className='flex items-center gap-3 truncate'>
              <span className='font-semibold'>Selected:</span>
              <span className='truncate'>{selectedMedia.filename}</span>
              <span className='text-muted-foreground'>
                ({Math.round(selectedMedia.size_bytes / 1024)} KB)
              </span>
            </div>
            <span className='text-muted-foreground uppercase font-mono'>
              {selectedMedia.mime_type}
            </span>
          </div>
        )}

        <DialogFooter className='pt-2 border-t flex justify-end gap-2'>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button disabled={!selectedMedia} onClick={handleConfirm}>
            Use Selected Image
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
