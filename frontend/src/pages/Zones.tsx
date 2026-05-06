import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Globe } from 'lucide-react'
import { zonesApi } from '../api/client'
import { Card, CardContent, CardHeader, CardTitle } from '../components/Card'
import { Button } from '../components/Button'
import { Input } from '../components/Input'

interface Zone {
  name: string
  type: string
}

export default function Zones() {
  const queryClient = useQueryClient()
  const [showAddZone, setShowAddZone] = useState(false)
  const [showAddData, setShowAddData] = useState(false)
  const [newZoneName, setNewZoneName] = useState('')
  const [newZoneType, setNewZoneType] = useState('static')
  const [newData, setNewData] = useState('')

  const { data: zones = [], isLoading: zonesLoading } = useQuery<Zone[]>({
    queryKey: ['zones'],
    queryFn: async () => {
      const { data } = await zonesApi.list()
      return data || []
    },
  })

  const { data: zoneData = [], isLoading: dataLoading } = useQuery<string[]>({
    queryKey: ['zoneData'],
    queryFn: async () => {
      const { data } = await zonesApi.listData()
      return data || []
    },
  })

  const addZoneMutation = useMutation({
    mutationFn: () => zonesApi.add(newZoneName, newZoneType),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['zones'] })
      setNewZoneName('')
      setShowAddZone(false)
    },
  })

  const deleteZoneMutation = useMutation({
    mutationFn: (name: string) => zonesApi.remove(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['zones'] })
    },
  })

  const addDataMutation = useMutation({
    mutationFn: () => zonesApi.addData(newData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['zoneData'] })
      setNewData('')
      setShowAddData(false)
    },
  })

  const deleteDataMutation = useMutation({
    mutationFn: (name: string) => zonesApi.removeData(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['zoneData'] })
    },
  })

  if (zonesLoading || dataLoading) {
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
          <h1 className="text-2xl font-bold text-foreground">Zone Management</h1>
          <p className="text-sm text-muted-foreground">Manage local zones and DNS records</p>
        </div>
      </div>

      {/* Local Zones */}
      <Card className="mb-6">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Local Zones</CardTitle>
            <Button size="sm" onClick={() => setShowAddZone(!showAddZone)}>
              <Plus className="h-4 w-4 mr-2" />
              Add Zone
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {showAddZone && (
            <div className="mb-4 p-4 bg-muted rounded-lg">
              <div className="flex gap-3 items-end">
                <div className="flex-1">
                  <Input
                    label="Zone Name"
                    value={newZoneName}
                    onChange={(e) => setNewZoneName(e.target.value)}
                    placeholder="example.local."
                  />
                </div>
                <div className="w-48">
                  <label className="text-sm font-medium text-foreground">Type</label>
                  <select
                    value={newZoneType}
                    onChange={(e) => setNewZoneType(e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                  >
                    <option value="static">static</option>
                    <option value="transparent">transparent</option>
                    <option value="redirect">redirect</option>
                    <option value="deny">deny</option>
                    <option value="refuse">refuse</option>
                    <option value="always_refuse">always_refuse</option>
                    <option value="always_nxdomain">always_nxdomain</option>
                    <option value="nodefault">nodefault</option>
                  </select>
                </div>
                <Button
                  size="sm"
                  onClick={() => addZoneMutation.mutate()}
                  disabled={!newZoneName || addZoneMutation.isPending}
                >
                  Add
                </Button>
              </div>
            </div>
          )}

          {zones.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Globe className="h-12 w-12 mx-auto mb-3 opacity-50" />
              <p>No local zones configured</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left py-3 px-4 font-medium text-muted-foreground">Zone Name</th>
                    <th className="text-left py-3 px-4 font-medium text-muted-foreground">Type</th>
                    <th className="text-right py-3 px-4 font-medium text-muted-foreground">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {zones.map((zone) => (
                    <tr key={zone.name} className="border-b border-border/50 hover:bg-muted/50">
                      <td className="py-3 px-4 font-mono">{zone.name}</td>
                      <td className="py-3 px-4">
                        <span className="px-2 py-1 rounded-full text-xs bg-primary/10 text-primary font-medium">
                          {zone.type}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => deleteZoneMutation.mutate(zone.name)}
                          disabled={deleteZoneMutation.isPending}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Local Data */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Local Data (DNS Records)</CardTitle>
            <Button size="sm" onClick={() => setShowAddData(!showAddData)}>
              <Plus className="h-4 w-4 mr-2" />
              Add Record
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {showAddData && (
            <div className="mb-4 p-4 bg-muted rounded-lg">
              <div className="flex gap-3 items-end">
                <div className="flex-1">
                  <Input
                    label="Record Data"
                    value={newData}
                    onChange={(e) => setNewData(e.target.value)}
                    placeholder="myhost.local. IN A 192.168.1.100"
                  />
                </div>
                <Button
                  size="sm"
                  onClick={() => addDataMutation.mutate()}
                  disabled={!newData || addDataMutation.isPending}
                >
                  Add
                </Button>
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                Format: name TTL class type data (e.g., "myhost.local. 3600 IN A 192.168.1.100")
              </p>
            </div>
          )}

          {zoneData.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No local data records</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left py-3 px-4 font-medium text-muted-foreground">Record</th>
                    <th className="text-right py-3 px-4 font-medium text-muted-foreground">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {zoneData.map((record, index) => (
                    <tr key={index} className="border-b border-border/50 hover:bg-muted/50">
                      <td className="py-3 px-4 font-mono text-xs">{record}</td>
                      <td className="py-3 px-4 text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            const name = record.split(/\s+/)[0]
                            deleteDataMutation.mutate(name)
                          }}
                          disabled={deleteDataMutation.isPending}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
