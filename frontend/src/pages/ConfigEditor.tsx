import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Save, RotateCcw, CheckCircle, AlertCircle } from 'lucide-react'
import { configApi } from '../api/client'
import { Card, CardContent, CardHeader, CardTitle } from '../components/Card'
import { Button } from '../components/Button'

export default function ConfigEditor() {
  const queryClient = useQueryClient()
  const [editedConfig, setEditedConfig] = useState<string | null>(null)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const { data: config, isLoading } = useQuery({
    queryKey: ['config'],
    queryFn: async () => {
      const { data } = await configApi.get()
      return data
    },
  })

  const saveMutation = useMutation({
    mutationFn: (raw: string) => configApi.update(raw),
    onSuccess: () => {
      setMessage({ type: 'success', text: 'Configuration saved successfully' })
      queryClient.invalidateQueries({ queryKey: ['config'] })
      setEditedConfig(null)
      setTimeout(() => setMessage(null), 3000)
    },
    onError: (err: Error) => {
      setMessage({ type: 'error', text: `Failed to save: ${err.message}` })
    },
  })

  const reloadMutation = useMutation({
    mutationFn: () => configApi.reload(),
    onSuccess: () => {
      setMessage({ type: 'success', text: 'Unbound reloaded successfully' })
      setTimeout(() => setMessage(null), 3000)
    },
    onError: (err: Error) => {
      setMessage({ type: 'error', text: `Failed to reload: ${err.message}` })
    },
  })

  const validateMutation = useMutation({
    mutationFn: () => configApi.validate(),
    onSuccess: ({ data }) => {
      if (data.valid) {
        setMessage({ type: 'success', text: 'Configuration is valid' })
      } else {
        setMessage({ type: 'error', text: `Invalid: ${data.message}` })
      }
      setTimeout(() => setMessage(null), 5000)
    },
  })

  const currentConfig = editedConfig ?? config?.raw ?? ''
  const hasChanges = editedConfig !== null && editedConfig !== config?.raw

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Configuration</h1>
          <p className="text-sm text-muted-foreground">Edit unbound.conf directly</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => validateMutation.mutate()}
            disabled={validateMutation.isPending}
          >
            <CheckCircle className="h-4 w-4 mr-2" />
            Validate
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => reloadMutation.mutate()}
            disabled={reloadMutation.isPending}
          >
            <RotateCcw className="h-4 w-4 mr-2" />
            Reload Unbound
          </Button>
          <Button
            size="sm"
            onClick={() => saveMutation.mutate(currentConfig)}
            disabled={!hasChanges || saveMutation.isPending}
          >
            <Save className="h-4 w-4 mr-2" />
            {saveMutation.isPending ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>

      {message && (
        <div
          className={`mb-4 px-4 py-3 rounded-md flex items-center gap-2 text-sm ${
            message.type === 'success'
              ? 'bg-success/10 text-success'
              : 'bg-destructive/10 text-destructive'
          }`}
        >
          {message.type === 'success' ? (
            <CheckCircle className="h-4 w-4" />
          ) : (
            <AlertCircle className="h-4 w-4" />
          )}
          {message.text}
        </div>
      )}

      {/* Parsed Config Summary */}
      {config?.server && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <Card>
            <CardContent className="py-4">
              <p className="text-sm text-muted-foreground">Port</p>
              <p className="text-lg font-semibold">{config.server.port}</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="py-4">
              <p className="text-sm text-muted-foreground">Interfaces</p>
              <p className="text-lg font-semibold">{config.server.interface?.length || 0}</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="py-4">
              <p className="text-sm text-muted-foreground">Forward Zones</p>
              <p className="text-lg font-semibold">{config.forward_zones?.length || 0}</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Raw Editor */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Raw Configuration</CardTitle>
            {hasChanges && (
              <span className="text-xs text-warning font-medium">Unsaved changes</span>
            )}
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <textarea
            value={currentConfig}
            onChange={(e) => setEditedConfig(e.target.value)}
            className="w-full h-[500px] p-4 font-mono text-sm bg-gray-900 text-gray-100 border-0 rounded-b-lg resize-none focus:outline-none focus:ring-0"
            spellCheck={false}
          />
        </CardContent>
      </Card>
    </div>
  )
}
