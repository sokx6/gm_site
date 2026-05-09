import apiClient from './client'

export interface PendingUser {
  id: number
  email: string
  nickname: string
  role: string
  status: string
  created_at: string
  updated_at: string
}

export interface PendingUsersResponse {
  data: PendingUser[]
}

export interface AdminActionResponse {
  message: string
}

export async function getPendingUsers() {
  const res = await apiClient.get<PendingUsersResponse>('/api/admin/users/pending')
  return res.data
}

export async function approveUser(id: number) {
  const res = await apiClient.put<AdminActionResponse>(`/api/admin/users/${id}/approve`)
  return res.data
}

export async function rejectUser(id: number) {
  const res = await apiClient.put<AdminActionResponse>(`/api/admin/users/${id}/reject`)
  return res.data
}
