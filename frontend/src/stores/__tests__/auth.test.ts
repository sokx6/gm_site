import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '../auth'

// Mock the auth API module
vi.mock('@/api/auth', () => ({
  login: vi.fn(),
  register: vi.fn(),
}))

import * as authApi from '@/api/auth'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('login sets accessToken and refreshToken', async () => {
    const mockRes = {
      data: {
        access_token: 'test-access-token',
        refresh_token: 'test-refresh-token',
        user: { id: 1, email: 'test@test.com', nickname: 'Tester', role: 'user', status: 'approved' },
      },
    }
    vi.mocked(authApi.login).mockResolvedValue(mockRes as never)

    const store = useAuthStore()
    await store.login('test@test.com', 'password123')

    expect(store.accessToken).toBe('test-access-token')
    expect(store.refreshToken).toBe('test-refresh-token')
    expect(store.user).toEqual(mockRes.data.user)
    expect(localStorage.getItem('accessToken')).toBe('test-access-token')
    expect(localStorage.getItem('refreshToken')).toBe('test-refresh-token')
  })

  it('login sets user from response', async () => {
    const mockUser = { id: 42, email: 'alice@test.com', nickname: 'Alice', role: 'admin', status: 'approved' }
    vi.mocked(authApi.login).mockResolvedValue({
      data: {
        access_token: 'tok',
        refresh_token: 'rtok',
        user: mockUser,
      },
    } as never)

    const store = useAuthStore()
    await store.login('alice@test.com', 'secret')

    expect(store.user).toEqual(mockUser)
  })

  it('login sets user with fallback when user not in response', async () => {
    vi.mocked(authApi.login).mockResolvedValue({
      data: {
        access_token: 'tok',
        refresh_token: 'rtok',
      },
    } as never)

    const store = useAuthStore()
    await store.login('test@test.com', 'pass')

    expect(store.user).toEqual({ id: 0, email: 'test@test.com', nickname: '', role: 'user', status: 'approved' })
  })

  it('logout clears tokens and user', () => {
    localStorage.setItem('accessToken', 'test-token')
    localStorage.setItem('refreshToken', 'test-refresh')

    const store = useAuthStore()
    store.accessToken = 'test-token'
    store.refreshToken = 'test-refresh'
    store.user = { id: 1, email: 'a@b.com', nickname: 'A', role: 'user', status: 'approved' }

    store.logout()

    expect(store.accessToken).toBe('')
    expect(store.refreshToken).toBe('')
    expect(store.user).toBeNull()
    expect(localStorage.getItem('accessToken')).toBeNull()
    expect(localStorage.getItem('refreshToken')).toBeNull()
  })

  it('isAdmin is true when user role is admin', () => {
    const store = useAuthStore()
    expect(store.isAdmin).toBe(false)

    store.user = { id: 1, email: 'admin@test.com', nickname: 'Admin', role: 'admin', status: 'approved' }
    expect(store.isAdmin).toBe(true)
  })

  it('isAdmin is false when user role is not admin', () => {
    const store = useAuthStore()
    store.user = { id: 2, email: 'user@test.com', nickname: 'User', role: 'user', status: 'approved' }
    expect(store.isAdmin).toBe(false)
  })

  it('isAdmin is false when user is null', () => {
    const store = useAuthStore()
    store.user = null
    expect(store.isAdmin).toBe(false)
  })

  it('isLoggedIn is true when accessToken is set', () => {
    const store = useAuthStore()
    expect(store.isLoggedIn).toBe(false)

    store.accessToken = 'some-token'
    expect(store.isLoggedIn).toBe(true)
  })

  it('register calls API and returns result', async () => {
    const mockRes = { data: { message: '注册成功，请等待管理员审核' } }
    vi.mocked(authApi.register).mockResolvedValue(mockRes as never)

    const store = useAuthStore()
    const result = await store.register('new@test.com', 'pass123', 'NewUser')

    expect(authApi.register).toHaveBeenCalledWith('new@test.com', 'pass123', 'NewUser')
    expect(result).toEqual(mockRes)
  })

  it('tryRestoreSession decodes JWT and sets user', () => {
    const store = useAuthStore()

    // Valid JWT: header.payload.signature
    // payload: { "user_id": 5, "role": "admin", ... }
    const payload = btoa(JSON.stringify({ user_id: 5, role: 'admin', exp: 9999999999 }))
    const token = `header.${payload}.signature`

    store.accessToken = token
    store.tryRestoreSession()

    expect(store.user).toEqual({ id: 5, email: '', nickname: '', role: 'admin', status: 'approved' })
  })

  it('tryRestoreSession calls logout on invalid token', () => {
    const store = useAuthStore()
    store.accessToken = 'invalid-token-no-dots'
    store.user = { id: 1, email: 'x@y.com', nickname: 'X', role: 'user', status: 'approved' }

    store.tryRestoreSession()

    expect(store.user).toBeNull()
    expect(store.accessToken).toBe('')
  })

  it('tryRestoreSession does nothing when no accessToken', () => {
    const store = useAuthStore()
    store.accessToken = ''

    store.tryRestoreSession()

    expect(store.user).toBeNull()
    expect(store.accessToken).toBe('')
  })
})
