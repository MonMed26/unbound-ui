import { useQuery } from '@tanstack/react-query'
import { Activity, Database, Clock, Cpu, RefreshCw, Trash2 } from 'lucide-react'
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { statsApi, cacheApi } from '../api/client'
import { Card, CardContent, CardHeader, CardTitle } from '../components/Card'
import { Button } from '../components/Button'
import { formatUptime, formatNumber } from '../lib/utils'

interface Stats {
  total_queries: number
  cache_hits: number
  cache_misses: number
  cache_hit_ratio: number
  uptime: number
  memory_mb: number
  num_threads: number
  query_types: Record<string, number>
  response_codes: Record<string, number>
  avg_recursion_ms: number
}

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

export default function Dashboard() {
  const { data: stats, isLoading, refetch } = useQuery<Stats>({
    queryKey: ['stats'],
    queryFn: async () => {
      const { data } = await statsApi.get()
      return data
    },
    refetchInterval: 5000,
  })

  const handleFlushCache = async () => {
    try {
      await cacheApi.flush()
      refetch()
    } catch (err) {
      console.error('Failed to flush cache:', err)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  const queryTypeData = stats?.query_types
    ? Object.entries(stats.query_types)
        .sort(([, a], [, b]) => b - a)
        .slice(0, 6)
        .map(([name, value]) => ({ name, value }))
    : []

  const responseCodeData = stats?.response_codes
    ? Object.entries(stats.response_codes)
        .map(([name, value]) => ({ name, value }))
    : []

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>
          <p className="text-sm text-muted-foreground">Unbound DNS server statistics</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button variant="destructive" size="sm" onClick={handleFlushCache}>
            <Trash2 className="h-4 w-4 mr-2" />
            Flush Cache
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="flex items-center gap-4 py-5">
            <div className="p-3 rounded-full bg-primary/10">
              <Activity className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Total Queries</p>
              <p className="text-2xl font-bold">{formatNumber(stats?.total_queries || 0)}</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex items-center gap-4 py-5">
            <div className="p-3 rounded-full bg-success/10">
              <Database className="h-5 w-5 text-success" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Cache Hit Ratio</p>
              <p className="text-2xl font-bold">{(stats?.cache_hit_ratio || 0).toFixed(1)}%</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex items-center gap-4 py-5">
            <div className="p-3 rounded-full bg-warning/10">
              <Clock className="h-5 w-5 text-warning" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Uptime</p>
              <p className="text-2xl font-bold">{formatUptime(stats?.uptime || 0)}</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex items-center gap-4 py-5">
            <div className="p-3 rounded-full bg-secondary/10">
              <Cpu className="h-5 w-5 text-secondary" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Memory Usage</p>
              <p className="text-2xl font-bold">{(stats?.memory_mb || 0).toFixed(1)} MB</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Query Types</CardTitle>
          </CardHeader>
          <CardContent>
            {queryTypeData.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie
                    data={queryTypeData}
                    cx="50%"
                    cy="50%"
                    outerRadius={80}
                    dataKey="value"
                    label={({ name, percent }) => `${name} (${((percent ?? 0) * 100).toFixed(0)}%)`}
                  >
                    {queryTypeData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[250px] flex items-center justify-center text-muted-foreground">
                No query data available
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Response Codes</CardTitle>
          </CardHeader>
          <CardContent>
            {responseCodeData.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <BarChart data={responseCodeData}>
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="value" fill="#3b82f6" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[250px] flex items-center justify-center text-muted-foreground">
                No response code data available
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Additional Info */}
      <div className="mt-6">
        <Card>
          <CardHeader>
            <CardTitle>Server Info</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <p className="text-sm text-muted-foreground">Cache Hits</p>
                <p className="text-lg font-semibold">{formatNumber(stats?.cache_hits || 0)}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Cache Misses</p>
                <p className="text-lg font-semibold">{formatNumber(stats?.cache_misses || 0)}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Avg Recursion</p>
                <p className="text-lg font-semibold">{(stats?.avg_recursion_ms || 0).toFixed(2)} ms</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Threads</p>
                <p className="text-lg font-semibold">{stats?.num_threads || 0}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
