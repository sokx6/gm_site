import apiClient from './client'

export interface ImageData {
  id: number
  title: string
  description: string
  url: string
  thumbnail_url: string
  album_id: number
  user_id: number
  width: number
  height: number
  file_size: number
  mime_type: string
  tags: string[]
  exif: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface ImageListParams {
  page?: number
  page_size?: number
  album_id?: number
  sort?: string
  order?: 'asc' | 'desc'
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ImageResponse {
  data: ImageData
}

export interface ImageListResponse {
  data: PaginatedResponse<ImageData>
}

export interface UpdateImageData {
  title?: string
  description?: string
  album_id?: number
  tags?: string[]
}

export async function uploadImage(formData: FormData) {
  const res = await apiClient.post<ImageResponse>('/api/images/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data
}

export async function getImages(params?: ImageListParams) {
  const res = await apiClient.get<ImageListResponse>('/api/images', { params })
  return res.data
}

export async function getImage(id: number) {
  const res = await apiClient.get<ImageResponse>(`/api/images/${id}`)
  return res.data
}

export async function updateImage(id: number, data: UpdateImageData) {
  const res = await apiClient.put<ImageResponse>(`/api/images/${id}`, data)
  return res.data
}

export async function deleteImage(id: number) {
  const res = await apiClient.delete<{ message: string }>(`/api/images/${id}`)
  return res.data
}

export async function searchImages(query: string, params?: ImageListParams) {
  const res = await apiClient.get<ImageListResponse>('/api/images/search', {
    params: { q: query, ...params },
  })
  return res.data
}
