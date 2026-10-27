import {
  LayoutDashboard,
  Newspaper,
  FilePlus,
  Image,
  FolderTree,
  Tag,
  UserCheck,
  ArrowLeftRight,
  Settings,
  Users,
  Building2,
} from 'lucide-react'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  user: {
    name: 'Editorial Admin',
    email: 'admin@newsportal.id',
    avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&auto=format&fit=crop&q=80',
  },
  teams: [
    {
      name: 'NewsPortal Editorial',
      logo: Building2,
      plan: 'Newsroom Pro',
    },
  ],
  navGroups: [
    {
      title: 'Editorial Newsroom',
      items: [
        {
          title: 'Dashboard',
          url: '/',
          icon: LayoutDashboard,
        },
        {
          title: 'All Articles',
          url: '/articles',
          icon: Newspaper,
        },
        {
          title: 'New Article',
          url: '/articles/new',
          icon: FilePlus,
        },
        {
          title: 'Media Library',
          url: '/media',
          icon: Image,
        },
      ],
    },
    {
      title: 'Taxonomy & Authors',
      items: [
        {
          title: 'Categories',
          url: '/categories',
          icon: FolderTree,
        },
        {
          title: 'Tags',
          url: '/tags',
          icon: Tag,
        },
        {
          title: 'Authors',
          url: '/authors',
          icon: UserCheck,
        },
      ],
    },
    {
      title: 'Site Management',
      items: [
        {
          title: '301 Redirects',
          url: '/redirects',
          icon: ArrowLeftRight,
        },
        {
          title: 'Team Users',
          url: '/users',
          icon: Users,
        },
        {
          title: 'Settings',
          url: '/settings',
          icon: Settings,
        },
      ],
    },
  ],
}
