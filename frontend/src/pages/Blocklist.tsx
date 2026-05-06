import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Shield, RefreshCw, ToggleLeft, ToggleRight } from 'lucide-react'
import { blocklistApi } from '../api/client'
import { Card, CardContent, CardHeader, CardTitle } from '../components/Card'
import { Button } from '../components/Button'
import { Input } from '../components/Input'

interface Source {
  id: string
  name: string
  url: string
  enabled: boolean
  count: number
  updated_at: string
}

interface BlocklistStats {
  total_sources: number
  enabled_sources: number
  total_domains: number
  manual_blocks: number
  whitelisted: number
}

export default function Blocklist() {
  const queryClient = useQueryClient()
  const [showAddSource, setShowAddSource] = useState(false)
  const [showBlockDomain, setShowBlockDomain] = useState(false)
  const [showWhitelist, setShowWhitelist] = useState(false)
  const [newSourceName, setNewSourceName] = useState('')
  const [newSourceUrl, setNewSourceUrl] = useState('')
  const [blockDomain, setBlockDomain] = useState('')
  const [whitelistDomain, setWhitelistDomain] = useState('')
  const [activeTab, setActiveTab] = useState<'sources' | 'manual' | 'whitelist'>('sources')

  const { data: sources = [] } = useQuery<Source[]>({
    queryKey: ['blocklist-sources'],
    queryFn: async () => {
      const { data } = await blocklistApi.getSources()
      return data
    },
  })

  const { data: stats } = useQuery<BlocklistStats>({
    queryKey: ['blocklist-stats'],
    queryFn: async () => {
      const { data } = await blocklistApi.getStats()
      return data
    },
  })

  const { data: manualBlocks = [] } = useQuery<string[]>({
    queryKey: ['manual-blocks'],
    queryFn: async () => {
      const { data } = await blocklistApi.getManualBlocks()
      return data || []
    },
  })

  const { data: whitelist = [] } = useQuery<string[]>({
    queryKey: ['whitelist'],
    queryFn: async () => {
      const { data } = await blocklistApi.getWhitelist()
      return data || []
    },
  })

  const addSourceMutation = useMutation({
    mutationFn: () => blocklistApi.addSource(newSourceName, newSourceUrl),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['blocklist-sources'] })
      setNewSourceName('')
      setNewSourceUrl('')
      setShowAddSource(false)
    },
  })

  const removeSourceMutation = useMutation({
    mutationFn: (id: string) => blocklistApi.removeSource(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['blocklist-sources'] })
      queryClient.invalidateQueries({ queryKey: ['blocklist-stats'] })
    },
  })

  const toggleSourceMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      blocklistApi.toggleSource(id, enabled),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['blocklist-sources'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: () => blocklistApi.update(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['blocklist-sources'] })
      queryClient.invalidateQueries({ queryKey: ['blocklist-stats'] })
    },
  })

  const blockMutation = useMutation({
    mutationFn: (domain: string) => blocklistApi.block(domain),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['manual-blocks'] })
      queryClient.invalidateQueries({ queryKey: ['blocklist-stats'] })
      setBlockDomain('')
      setShowBlockDomain(false)
    },
  })

  const unblockMutation = useMutation({
    mutationFn: (domain: string) => blocklistApi.unblock(domain),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['manual-blocks'] })
      queryClient.invalidateQueries({ queryKey: ['blocklist-stats'] })
    },
  })

  const addWhitelistMutation = useMutation({
    mutationFn: (domain: string) => blocklistApi.addWhitelist(domain),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['whitelist'] })
      queryClient.invalidateQueries({ queryKey: ['blocklist-stats'] })
      setWhitelistDomain('')
      setShowWhitelist(false)
    },
  })

  const removeWhitelistMutation = useMutation({
    mutationFn: (domain: string) => blocklistApi.removeWhitelist(domain),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['whitelist'] })
      queryClient.invalidateQueries({ queryKey: ['blocklist-stats'] })
    },
  })

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-foreground">DNS Blocklist</h1>
          <p className="text-sm text-muted-foreground">Block ads, trackers, and malware domains</p>
        </div>
        <Button
          size="sm"
          onClick={() => updateMutation.mutate()}
          disabled={updateMutation.isPending}
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${updateMutation.isPending ? 'animate-spin' : ''}`} />
          {updateMutation.isPending ? 'Updating...' : 'Update All'}
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="py-4 text-center">
            <p className="text-2xl font-bold text-primary">{stats?.total_domains || 0}</p>
            <p className="text-xs text-muted-foreground">Blocked Domains</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4 text-center">
            <p className="text-2xl font-bold">{stats?.enabled_sources || 0}</p>
            <p className="text-xs text-muted-foreground">Active Sources</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4 text-center">
            <p className="text-2xl font-bold">{stats?.manual_blocks || 0}</p>
            <p className="text-xs text-muted-foreground">Manual Blocks</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-4 text-center">
            <p className="text-2xl font-bold">{stats?.whitelisted || 0}</p>
            <p className="text-xs text-muted-foreground">Whitelisted</p>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-4 bg-muted p-1 rounded-lg w-fit">
        {(['sources', 'manual', 'whitelist'] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab
                ? 'bg-white text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {tab === 'sources' ? 'Sources' : tab === 'manual' ? 'Manual Blocks' : 'Whitelist'}
          </button>
        ))}
      </div>

      {/* Sources Tab */}
      {activeTab === 'sources' && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Blocklist Sources</CardTitle>
              <Button size="sm" onClick={() => setShowAddSource(!showAddSource)}>
                <Plus className="h-4 w-4 mr-2" />
                Add Source
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {showAddSource && (
              <div className="mb-4 p-4 bg-muted rounded-lg space-y-3">
                <Input
                  label="Source Name"
                  value={newSourceName}
                  onChange={(e) => setNewSourceName(e.target.value)}
                  placeholder="Steven Black Hosts"
                />
                <Input
                  label="URL"
                  value={newSourceUrl}
                  onChange={(e) => setNewSourceUrl(e.target.value)}
                  placeholder="https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"
                />
                <Button
                  size="sm"
                  onClick={() => addSourceMutation.mutate()}
                  disabled={!newSourceName || !newSourceUrl || addSourceMutation.isPending}
                >
                  Add Source
                </Button>
              </div>
            )}

            {sources.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <Shield className="h-12 w-12 mx-auto mb-3 opacity-50" />
                <p>No blocklist sources configured</p>
                <p className="text-xs mt-1">Add a source to start blocking domains</p>
              </div>
            ) : (
              <div className="space-y-3">
                {sources.map((source) => (
                  <div
                    key={source.id}
                    className="flex items-center justify-between p-4 border border-border rounded-lg"
                  >
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium text-sm">{source.name}</h4>
                        <span
                          className={`px-2 py-0.5 rounded-full text-xs ${
                            source.enabled
                              ? 'bg-success/10 text-success'
                              : 'bg-muted text-muted-foreground'
                          }`}
                        >
                          {source.enabled ? 'Active' : 'Disabled'}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground mt-1 truncate max-w-md">
                        {source.url}
                      </p>
                      <p className="text-xs text-muted-foreground mt-1">
                        {source.count > 0 ? `${source.count.toLocaleString()} domains` : 'Not yet updated'}
                        {source.updated_at && ` • Last updated: ${new Date(source.updated_at).toLocaleDateString()}`}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() =>
                          toggleSourceMutation.mutate({ id: source.id, enabled: !source.enabled })
                        }
                        className="text-muted-foreground hover:text-foreground"
                      >
                        {source.enabled ? (
                          <ToggleRight className="h-6 w-6 text-success" />
                        ) : (
                          <ToggleLeft className="h-6 w-6" />
                        )}
                      </button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => removeSourceMutation.mutate(source.id)}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Manual Blocks Tab */}
      {activeTab === 'manual' && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Manually Blocked Domains</CardTitle>
              <Button size="sm" onClick={() => setShowBlockDomain(!showBlockDomain)}>
                <Plus className="h-4 w-4 mr-2" />
                Block Domain
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {showBlockDomain && (
              <div className="mb-4 p-4 bg-muted rounded-lg">
                <div className="flex gap-3 items-end">
                  <div className="flex-1">
                    <Input
                      label="Domain"
                      value={blockDomain}
                      onChange={(e) => setBlockDomain(e.target.value)}
                      placeholder="ads.example.com"
                    />
                  </div>
                  <Button
                    size="sm"
                    onClick={() => blockMutation.mutate(blockDomain)}
                    disabled={!blockDomain || blockMutation.isPending}
                  >
                    Block
                  </Button>
                </div>
              </div>
            )}

            {manualBlocks.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <p>No manually blocked domains</p>
              </div>
            ) : (
              <div className="space-y-2">
                {manualBlocks.map((domain) => (
                  <div
                    key={domain}
                    className="flex items-center justify-between py-2 px-4 border border-border rounded-md"
                  >
                    <span className="font-mono text-sm">{domain}</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => unblockMutation.mutate(domain)}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Whitelist Tab */}
      {activeTab === 'whitelist' && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Whitelisted Domains</CardTitle>
              <Button size="sm" onClick={() => setShowWhitelist(!showWhitelist)}>
                <Plus className="h-4 w-4 mr-2" />
                Add to Whitelist
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {showWhitelist && (
              <div className="mb-4 p-4 bg-muted rounded-lg">
                <div className="flex gap-3 items-end">
                  <div className="flex-1">
                    <Input
                      label="Domain"
                      value={whitelistDomain}
                      onChange={(e) => setWhitelistDomain(e.target.value)}
                      placeholder="allowed.example.com"
                    />
                  </div>
                  <Button
                    size="sm"
                    onClick={() => addWhitelistMutation.mutate(whitelistDomain)}
                    disabled={!whitelistDomain || addWhitelistMutation.isPending}
                  >
                    Add
                  </Button>
                </div>
              </div>
            )}

            {whitelist.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <p>No whitelisted domains</p>
                <p className="text-xs mt-1">Whitelisted domains will bypass blocklist filtering</p>
              </div>
            ) : (
              <div className="space-y-2">
                {whitelist.map((domain) => (
                  <div
                    key={domain}
                    className="flex items-center justify-between py-2 px-4 border border-border rounded-md"
                  >
                    <span className="font-mono text-sm">{domain}</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => removeWhitelistMutation.mutate(domain)}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
