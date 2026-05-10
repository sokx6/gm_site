import apiClient from './client'

export interface AlbumData {
  id: number
  name: string
  description: string
  cover_url: string
  user_id: number
  image_count: number
  privacy?: string
  created_at: string
  updated_at: string
}

export interface AlbumListResponse {
  data: AlbumData[]
}

export interface AlbumResponse {
  data: AlbumData
}

export interface CreateAlbumData {
  name: string
  description?: string
  cover_url?: string
}

export async function getAlbums() {
  const res = await apiClient.get<AlbumListResponse>('/api/albums')
  return res.data
}

export async function createAlbum(data: CreateAlbumData) {
  const res = await apiClient.post<AlbumResponse>('/api/albums', data)
  return res.data
}

export interface UpdateAlbumData {
  name?: string
  description?: string
}

export async function updateAlbum(id: number, data: UpdateAlbumData) {
  const res = await apiClient.put<AlbumResponse>(`/api/albums/${id}`, data)
  return res.data
}

export async function deleteAlbum(id: number) {
  const res = await apiClient.delete<{ message: string; deleted?: boolean }>(`/api/albums/${id}`)
  return res.data
}
