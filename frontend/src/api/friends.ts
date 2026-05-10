import apiClient from './client'

// ── Types ──────────────────────────────────────────────────
export interface FriendRequest {
  id: number
  from_user_id: number
  to_user_id: number
  status: string // pending | accepted | rejected
  created_at: string
  updated_at: string
  from_nickname?: string
  from_email?: string
  to_nickname?: string
  to_email?: string
}

export interface Friend {
  id: number
  user_id: number
  friend_id: number
  created_at: string
  friend_nickname?: string
  friend_email?: string
}

export interface GetRequestsResponse {
  code: number
  message: string
  data: FriendRequest[]
}

export interface GetFriendsResponse {
  code: number
  message: string
  data: Friend[]
}

export interface SendRequestResponse {
  code: number
  message: string
  data: FriendRequest
}

export interface SuccessResponse {
  code: number
  message: string
  data: null
}

// ── API functions ──────────────────────────────────────────

/** GET /api/friends/requests — pending incoming requests */
export async function getFriendRequests(): Promise<GetRequestsResponse> {
  const res = await apiClient.get<GetRequestsResponse>('/api/friends/requests')
  return res.data
}

/** GET /api/friends — friend list */
export async function getFriends(): Promise<GetFriendsResponse> {
  const res = await apiClient.get<GetFriendsResponse>('/api/friends')
  return res.data
}

/**
 * POST /api/friends/request — send friend request
 * @param email - Email of the user to send request to
 */
export async function sendFriendRequest(email: string): Promise<SendRequestResponse> {
  const res = await apiClient.post<SendRequestResponse>('/api/friends/request', {
    email,
  })
  return res.data
}

/** PUT /api/friends/request/:id/accept */
export async function acceptRequest(id: number): Promise<SuccessResponse> {
  const res = await apiClient.put<SuccessResponse>(`/api/friends/request/${id}/accept`)
  return res.data
}

/** PUT /api/friends/request/:id/reject */
export async function rejectRequest(id: number): Promise<SuccessResponse> {
  const res = await apiClient.put<SuccessResponse>(`/api/friends/request/${id}/reject`)
  return res.data
}

/** DELETE /api/friends/:id */
export async function deleteFriend(id: number): Promise<SuccessResponse> {
  const res = await apiClient.delete<SuccessResponse>(`/api/friends/${id}`)
  return res.data
}
