import apiClient from './client'

export interface User {
  id: number
  email: string
  nickname: string
  role: string
  status: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  nickname: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
}

export interface AuthResponse {
  data: TokenPair & { user: User }
}

export interface RefreshResponse {
  data: TokenPair
}

export async function register(email: string, password: string, nickname: string) {
  const res = await apiClient.post<AuthResponse>('/api/auth/register', {
    email,
    password,
    nickname,
  } satisfies RegisterRequest)
  return res.data
}

export async function login(email: string, password: string) {
  const res = await apiClient.post<AuthResponse>('/api/auth/login', {
    email,
    password,
  } satisfies LoginRequest)
  localStorage.setItem('accessToken', res.data.data.access_token)
  localStorage.setItem('refreshToken', res.data.data.refresh_token)
  return res.data
}

export async function refreshToken() {
  const refreshTokenValue = localStorage.getItem('refreshToken')
  const res = await apiClient.post<RefreshResponse>('/api/auth/refresh', {
    refresh_token: refreshTokenValue,
  })
  localStorage.setItem('accessToken', res.data.data.access_token)
  localStorage.setItem('refreshToken', res.data.data.refresh_token)
  return res.data
}

export function logout() {
  localStorage.removeItem('accessToken')
  localStorage.removeItem('refreshToken')
  window.location.hash = '#/login'
}
