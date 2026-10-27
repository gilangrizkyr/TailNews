import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import Youtube from '@tiptap/extension-youtube'
import Placeholder from '@tiptap/extension-placeholder'
import {
  Bold,
  Italic,
  Strikethrough,
  Heading2,
  Heading3,
  List,
  ListOrdered,
  Quote,
  Undo,
  Redo,
  Link as LinkIcon,
  Image as ImageIcon,
  Video as VideoIcon,
  Minus,
  RemoveFormatting,
} from 'lucide-react'
import { Button } from '@/components/ui/button'

interface TipTapEditorProps {
  content: string
  onChange: (html: string) => void
  placeholder?: string
  onOpenMediaPicker?: () => void
}

export function TipTapEditor({
  content,
  onChange,
  placeholder = 'Write the article story here...',
  onOpenMediaPicker,
}: TipTapEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: {
          levels: [2, 3],
        },
      }),
      Placeholder.configure({
        placeholder,
      }),
      Link.configure({
        openOnClick: false,
        HTMLAttributes: {
          class: 'text-primary underline font-medium hover:opacity-80',
        },
      }),
      Image.configure({
        inline: false,
        HTMLAttributes: {
          class: 'rounded-lg max-w-full my-4 shadow-sm border',
        },
      }),
      Youtube.configure({
        width: 640,
        height: 360,
        HTMLAttributes: {
          class: 'aspect-video w-full rounded-lg my-4 shadow-sm',
        },
      }),
    ],
    content,
    editorProps: {
      attributes: {
        class:
          'prose dark:prose-invert max-w-none min-h-[360px] p-4 focus:outline-none text-foreground text-base leading-relaxed',
      },
    },
    onUpdate: ({ editor }) => {
      onChange(editor.getHTML())
    },
  })

  if (!editor) {
    return null
  }

  const setLink = () => {
    const previousUrl = editor.getAttributes('link').href
    const url = window.prompt('Enter link URL:', previousUrl)

    if (url === null) return
    if (url === '') {
      editor.chain().focus().extendMarkRange('link').unsetLink().run()
      return
    }

    editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
  }

  const addYoutubeVideo = () => {
    const url = window.prompt('Enter YouTube URL:')
    if (url) {
      editor.commands.setYoutubeVideo({
        src: url,
        width: 640,
        height: 360,
      })
    }
  }

  const addImagePrompt = () => {
    if (onOpenMediaPicker) {
      onOpenMediaPicker()
      return
    }
    const url = window.prompt('Enter image URL:')
    if (url) {
      editor.chain().focus().setImage({ src: url }).run()
    }
  }

  return (
    <div className='border rounded-lg overflow-hidden bg-background'>
      {/* Sticky Toolbar */}
      <div className='flex flex-wrap items-center gap-1 p-2 bg-muted/40 border-b'>
        <Button
          type='button'
          size='sm'
          variant={editor.isActive('bold') ? 'secondary' : 'ghost'}
          onClick={() => editor.chain().focus().toggleBold().run()}
          title='Bold'
          className='h-8 w-8 p-0'
        >
          <Bold className='h-4 w-4' />
        </Button>
        <Button
          type='button'
          size='sm'
          variant={editor.isActive('italic') ? 'secondary' : 'ghost'}
          onClick={() => editor.chain().focus().toggleItalic().run()}
          title='Italic'
          className='h-8 w-8 p-0'
        >
          <Italic className='h-4 w-4' />
        </Button>
        <Button
          type='button'
          size='sm'
          variant={editor.isActive('strike') ? 'secondary' : 'ghost'}
          onClick={() => editor.chain().focus().toggleStrike().run()}
          title='Strikethrough'
          className='h-8 w-8 p-0'
        >
          <Strikethrough className='h-4 w-4' />
        </Button>

        <div className='h-4 w-[1px] bg-border mx-1' />

        <Button
          type='button'
          size='sm'
          variant={editor.isActive('heading', { level: 2 }) ? 'secondary' : 'ghost'}
          onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
          title='Heading 2'
          className='h-8 w-8 p-0 font-bold'
        >
          <Heading2 className='h-4 w-4' />
        </Button>
        <Button
          type='button'
          size='sm'
          variant={editor.isActive('heading', { level: 3 }) ? 'secondary' : 'ghost'}
          onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
          title='Heading 3'
          className='h-8 w-8 p-0'
        >
          <Heading3 className='h-4 w-4' />
        </Button>

        <div className='h-4 w-[1px] bg-border mx-1' />

        <Button
          type='button'
          size='sm'
          variant={editor.isActive('bulletList') ? 'secondary' : 'ghost'}
          onClick={() => editor.chain().focus().toggleBulletList().run()}
          title='Bullet List'
          className='h-8 w-8 p-0'
        >
          <List className='h-4 w-4' />
        </Button>
        <Button
          type='button'
          size='sm'
          variant={editor.isActive('orderedList') ? 'secondary' : 'ghost'}
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
          title='Ordered List'
          className='h-8 w-8 p-0'
        >
          <ListOrdered className='h-4 w-4' />
        </Button>
        <Button
          type='button'
          size='sm'
          variant={editor.isActive('blockquote') ? 'secondary' : 'ghost'}
          onClick={() => editor.chain().focus().toggleBlockquote().run()}
          title='Blockquote'
          className='h-8 w-8 p-0'
        >
          <Quote className='h-4 w-4' />
        </Button>
        <Button
          type='button'
          size='sm'
          variant='ghost'
          onClick={() => editor.chain().focus().setHorizontalRule().run()}
          title='Horizontal Line'
          className='h-8 w-8 p-0'
        >
          <Minus className='h-4 w-4' />
        </Button>

        <div className='h-4 w-[1px] bg-border mx-1' />

        <Button
          type='button'
          size='sm'
          variant={editor.isActive('link') ? 'secondary' : 'ghost'}
          onClick={setLink}
          title='Insert Link'
          className='h-8 w-8 p-0'
        >
          <LinkIcon className='h-4 w-4' />
        </Button>
        <Button
          type='button'
          size='sm'
          variant='ghost'
          onClick={addImagePrompt}
          title='Insert Image'
          className='h-8 w-8 p-0'
        >
          <ImageIcon className='h-4 w-4' />
        </Button>
        <Button
          type='button'
          size='sm'
          variant='ghost'
          onClick={addYoutubeVideo}
          title='Embed YouTube Video'
          className='h-8 w-8 p-0'
        >
          <VideoIcon className='h-4 w-4' />
        </Button>
        <Button
          type='button'
          size='sm'
          variant='ghost'
          onClick={() => editor.chain().focus().unsetAllMarks().clearNodes().run()}
          title='Clear Formatting'
          className='h-8 w-8 p-0'
        >
          <RemoveFormatting className='h-4 w-4' />
        </Button>

        <div className='ms-auto flex items-center gap-1'>
          <Button
            type='button'
            size='sm'
            variant='ghost'
            onClick={() => editor.chain().focus().undo().run()}
            disabled={!editor.can().undo()}
            title='Undo'
            className='h-8 w-8 p-0'
          >
            <Undo className='h-4 w-4' />
          </Button>
          <Button
            type='button'
            size='sm'
            variant='ghost'
            onClick={() => editor.chain().focus().redo().run()}
            disabled={!editor.can().redo()}
            title='Redo'
            className='h-8 w-8 p-0'
          >
            <Redo className='h-4 w-4' />
          </Button>
        </div>
      </div>

      {/* Editor Content Area */}
      <EditorContent editor={editor} />
    </div>
  )
}
