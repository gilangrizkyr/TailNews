import { useState, useEffect } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import {
  Newspaper,
  Eye,
  FileCheck2,
  FileClock,
  FolderTree,
  Users,
  Image as ImageIcon,
  ExternalLink,
  FilePlus,
  UploadCloud,
  Loader2,
  TrendingUp,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import { dashboardApi, articleApi } from '@/lib/api-client'
import { DashboardStatsResponse, ArticleResponse } from '@/types/api'

export function Dashboard() {
  const navigate = useNavigate()
  const [stats, setStats] = useState<DashboardStatsResponse | null>(null)
  const [recentArticles, setRecentArticles] = useState<ArticleResponse[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const loadDashboard = async () => {
      try {
        const [statsData, articlesData] = await Promise.all([
          dashboardApi.getStats(),
          articleApi.list({ page: 1, page_size: 6 }),
        ])
        setStats(statsData)
        setRecentArticles(articlesData.data)
      } catch (err) {
        console.error('Failed to load dashboard data:', err)
      } finally {
        setIsLoading(false)
      }
    }
    loadDashboard()
  }, [])

  return (
    <>
      <Header>
        <div className='flex items-center gap-2 me-auto'>
          <h2 className='text-sm font-semibold tracking-tight'>
            NewsPortal Newsroom CMS
          </h2>
          <Badge variant='outline' className='text-[10px] bg-primary/10 text-primary border-primary/20'>
            Production SSR
          </Badge>
        </div>
        <ThemeSwitch />
        <ProfileDropdown />
      </Header>

      <Main>
        <div className='flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>Editorial Overview</h1>
            <p className='text-sm text-muted-foreground'>
              Real-time editorial status, readership metrics, and publishing pipeline.
            </p>
          </div>
          <div className='flex items-center gap-2'>
            <Button
              variant='outline'
              size='sm'
              onClick={() => window.open('http://127.0.0.1:4321', '_blank')}
            >
              <ExternalLink className='h-4 w-4 me-1.5' /> Visit Public Portal
            </Button>
            <Button size='sm' onClick={() => navigate({ to: '/articles/new' })}>
              <FilePlus className='h-4 w-4 me-1.5' /> Write Article
            </Button>
          </div>
        </div>

        {isLoading ? (
          <div className='flex flex-col items-center justify-center h-72 gap-2 text-muted-foreground'>
            <Loader2 className='h-8 w-8 animate-spin' />
            <p className='text-sm'>Fetching live editorial statistics...</p>
          </div>
        ) : (
          <div className='space-y-6'>
            {/* KPI Metrics */}
            <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Total Articles
                  </CardTitle>
                  <Newspaper className='h-4 w-4 text-muted-foreground' />
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>
                    {(stats?.total_articles ?? 0).toLocaleString()}
                  </div>
                  <p className='text-xs text-muted-foreground mt-1'>
                    Across {stats?.total_categories || 0} news categories
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Published Live
                  </CardTitle>
                  <FileCheck2 className='h-4 w-4 text-emerald-500' />
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold text-emerald-600 dark:text-emerald-400'>
                    {(stats?.published_articles ?? 0).toLocaleString()}
                  </div>
                  <p className='text-xs text-muted-foreground mt-1'>
                    Live on Astro SSR reader portal
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Drafts & In-Review
                  </CardTitle>
                  <FileClock className='h-4 w-4 text-amber-500' />
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold text-amber-600 dark:text-amber-400'>
                    {((stats?.draft_articles || 0) + (stats?.scheduled_articles || 0)).toLocaleString()}
                  </div>
                  <p className='text-xs text-muted-foreground mt-1'>
                    {stats?.scheduled_articles || 0} scheduled for auto-publishing
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-sm font-medium'>
                    Total Readership
                  </CardTitle>
                  <Eye className='h-4 w-4 text-blue-500' />
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold text-blue-600 dark:text-blue-400 flex items-center gap-1.5'>
                    <TrendingUp className='h-5 w-5' />
                    {(stats?.total_views ?? 0).toLocaleString()}
                  </div>
                  <p className='text-xs text-muted-foreground mt-1'>
                    Aggregated reader page views
                  </p>
                </CardContent>
              </Card>
            </div>

            {/* Quick Actions & Recent Articles */}
            <div className='grid grid-cols-1 lg:grid-cols-3 gap-6'>
              {/* Recent Published Feed */}
              <Card className='lg:col-span-2'>
                <CardHeader className='flex flex-row items-center justify-between'>
                  <div>
                    <CardTitle className='text-base'>Recent Editorial Activity</CardTitle>
                    <CardDescription className='text-xs'>
                      Latest news articles drafted or published by the team.
                    </CardDescription>
                  </div>
                  <Button variant='ghost' size='sm' asChild className='text-xs'>
                    <Link to='/articles'>View All Articles</Link>
                  </Button>
                </CardHeader>
                <CardContent className='divide-y divide-border p-0'>
                  {recentArticles.map((art) => (
                    <div
                      key={art.id}
                      className='p-4 flex items-center justify-between hover:bg-muted/30 transition-colors gap-3'
                    >
                      <div className='flex items-center gap-3 min-w-0'>
                        <div className='h-11 w-14 rounded overflow-hidden bg-muted border shrink-0'>
                          {art.cover_url ? (
                            <img
                              src={art.cover_url}
                              alt={art.title}
                              className='h-full w-full object-cover'
                            />
                          ) : (
                            <div className='h-full w-full flex items-center justify-center text-[9px] text-muted-foreground'>
                              News
                            </div>
                          )}
                        </div>
                        <div className='min-w-0'>
                          <Link
                            to='/articles/$id'
                            params={{ id: String(art.id) }}
                            className='font-semibold text-xs hover:underline line-clamp-1'
                          >
                            {art.title}
                          </Link>
                          <div className='flex items-center gap-2 mt-1 text-[11px] text-muted-foreground'>
                            <span>{art.category_name}</span>
                            <span>•</span>
                            <span>{art.author_name}</span>
                            <span>•</span>
                            <span className='font-mono'>{art.view_count} views</span>
                          </div>
                        </div>
                      </div>

                      <Badge
                        variant='outline'
                        className={`text-[10px] capitalize shrink-0 ${
                          art.status === 'published'
                            ? 'bg-green-500/10 text-green-600 border-green-500/20'
                            : 'bg-muted text-muted-foreground'
                        }`}
                      >
                        {art.status}
                      </Badge>
                    </div>
                  ))}
                </CardContent>
              </Card>

              {/* Taxonomy Summary & Quick Links */}
              <div className='space-y-4'>
                <Card>
                  <CardHeader className='pb-3'>
                    <CardTitle className='text-base'>Editorial Directory</CardTitle>
                    <CardDescription className='text-xs'>
                      Content infrastructure entities
                    </CardDescription>
                  </CardHeader>
                  <CardContent className='space-y-3'>
                    <div className='flex items-center justify-between p-2 rounded-lg bg-muted/30'>
                      <div className='flex items-center gap-2.5 text-xs font-medium'>
                        <FolderTree className='h-4 w-4 text-primary' />
                        Categories
                      </div>
                      <span className='font-mono text-xs font-bold'>
                        {stats?.total_categories || 0}
                      </span>
                    </div>

                    <div className='flex items-center justify-between p-2 rounded-lg bg-muted/30'>
                      <div className='flex items-center gap-2.5 text-xs font-medium'>
                        <Users className='h-4 w-4 text-primary' />
                        Author By-lines
                      </div>
                      <span className='font-mono text-xs font-bold'>
                        {stats?.total_authors || 0}
                      </span>
                    </div>

                    <div className='flex items-center justify-between p-2 rounded-lg bg-muted/30'>
                      <div className='flex items-center gap-2.5 text-xs font-medium'>
                        <ImageIcon className='h-4 w-4 text-primary' />
                        Media Assets
                      </div>
                      <span className='font-mono text-xs font-bold'>
                        {stats?.total_media || 0}
                      </span>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className='pb-3'>
                    <CardTitle className='text-base'>Editorial Shortcuts</CardTitle>
                  </CardHeader>
                  <CardContent className='space-y-2'>
                    <Button
                      variant='outline'
                      size='sm'
                      className='w-full justify-start text-xs'
                      onClick={() => navigate({ to: '/media' })}
                    >
                      <UploadCloud className='h-4 w-4 me-2' /> Open Media Library
                    </Button>
                    <Button
                      variant='outline'
                      size='sm'
                      className='w-full justify-start text-xs'
                      onClick={() => navigate({ to: '/categories' })}
                    >
                      <FolderTree className='h-4 w-4 me-2' /> Manage Categories
                    </Button>
                    <Button
                      variant='outline'
                      size='sm'
                      className='w-full justify-start text-xs'
                      onClick={() => navigate({ to: '/redirects' })}
                    >
                      <ExternalLink className='h-4 w-4 me-2' /> View 301 SEO Redirects
                    </Button>
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        )}
      </Main>
    </>
  )
}
