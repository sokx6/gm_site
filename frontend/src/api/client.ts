import axios from 'axios'

const apiClient = axios.create({ 
  baseURL: import.meta.env.VITE_API_BASE_URL || '/' 
})

// Request interceptor: attach access token
apiClient.interceptors.request.use(config => {
  const token = localStorage.getItem('accessToken')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Response interceptor: handle 401 → refresh → retry
let isRefreshing = false
let failedQueue: Array<{ resolve: (token: string) => void; reject: (error: unknown) => void }> = []

apiClient.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise<string>((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        }).then(token => {
          originalRequest.headers.Authorization = `Bearer ${token}`
          return apiClient(originalRequest)
        })
      }
      originalRequest._retry = true
      isRefreshing = true
      try {
        const refreshToken = localStorage.getItem('refreshToken')
        const { data } = await axios.post('/api/auth/refresh', { refresh_token: refreshToken })
        localStorage.setItem('accessToken', data.data.access_token)
        localStorage.setItem('refreshToken', data.data.refresh_token)
        failedQueue.forEach(({ resolve }) => resolve(data.data.access_token))
        failedQueue = []
        originalRequest.headers.Authorization = `Bearer ${data.data.access_token}`
        return apiClient(originalRequest)
      } catch (refreshError) {
        failedQueue.forEach(({ reject }) => reject(refreshError))
        failedQueue = []
        localStorage.removeItem('accessToken')
        localStorage.removeItem('refreshToken')
        window.location.hash = '#/login'
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }
    return Promise.reject(error)
  }
)

export default apiClient
