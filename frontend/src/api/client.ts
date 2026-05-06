import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle 401 responses (skip for auth endpoints to let login page handle errors)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const url = error.config?.url || ''
    const isAuthEndpoint = url.includes('/auth/')
    if (error.response?.status === 401 && !isAuthEndpoint) {
      localStorage.removeItem('token')
      window.location.href = '/'
    }
    return Promise.reject(error)
  }
)

// Auth
export const authApi = {
  login: (username: string, password: string) =>
    api.post('/auth/login', { username, password }),
  setup: (username: string, password: string) =>
    api.post('/auth/setup', { username, password }),
  status: () => api.get('/auth/status'),
}

// Stats
export const statsApi = {
  get: () => api.get('/stats'),
}

// Config
export const configApi = {
  get: () => api.get('/config'),
  update: (raw: string) => api.put('/config', { raw }),
  validate: () => api.post('/config/validate'),
  reload: () => api.post('/config/reload'),
}

// Zones
export const zonesApi = {
  list: () => api.get('/zones'),
  add: (name: string, type: string) => api.post('/zones', { name, type }),
  remove: (name: string) => api.delete(`/zones/${encodeURIComponent(name)}`),
  listData: () => api.get('/zones/data'),
  addData: (data: string) => api.post('/zones/data', { data }),
  removeData: (name: string) => api.delete(`/zones/data/${encodeURIComponent(name)}`),
}

// Blocklist
export const blocklistApi = {
  getSources: () => api.get('/blocklist/sources'),
  addSource: (name: string, url: string) => api.post('/blocklist/sources', { name, url }),
  removeSource: (id: string) => api.delete(`/blocklist/sources/${id}`),
  toggleSource: (id: string, enabled: boolean) =>
    api.put(`/blocklist/sources/${id}/toggle`, { enabled }),
  update: () => api.post('/blocklist/update'),
  getStats: () => api.get('/blocklist/stats'),
  block: (domain: string) => api.post('/blocklist/block', { domain }),
  unblock: (domain: string) => api.post('/blocklist/unblock', { domain }),
  getManualBlocks: () => api.get('/blocklist/manual'),
  getWhitelist: () => api.get('/blocklist/whitelist'),
  addWhitelist: (domain: string) => api.post('/blocklist/whitelist', { domain }),
  removeWhitelist: (domain: string) => api.delete(`/blocklist/whitelist/${encodeURIComponent(domain)}`),
}

// Cache
export const cacheApi = {
  flush: () => api.post('/cache/flush'),
}

export default api
