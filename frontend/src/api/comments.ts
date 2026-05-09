import apiClient from './client'

export interface CommentData {
  id: number
  content: string
  user_id: number
  image_id: number
  nickname: string
  created_at: string
  updated_at: string
}

export interface CommentListResponse {
  data: CommentData[]
  total: number
  page: number
  page_size: number
}

export interface CommentResponse {
  data: CommentData
}

export interface CreateCommentData {
  content: string
}

export async function getComments(imageId: number, page: number = 1) {
  const res = await apiClient.get<CommentListResponse>(`/api/images/${imageId}/comments`, {
    params: { page },
  })
  return res.data
}

export async function createComment(imageId: number, content: string) {
  const res = await apiClient.post<CommentResponse>(`/api/images/${imageId}/comments`, {
    content,
  } satisfies CreateCommentData)
  return res.data
}

export async function deleteComment(id: number) {
  const res = await apiClient.delete<{ message: string }>(`/api/comments/${id}`)
  return res.data
}
