<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import * as friendsApi from '@/api/friends'
import type { FriendRequest, Friend } from '@/api/friends'

const auth = useAuthStore()
const isAuthenticated = computed(() => auth.isLoggedIn)

// ── State ──────────────────────────────────────────────────
const friendInput = ref('')
const sendLoading = ref(false)
const sendError = ref('')
const sendSuccess = ref('')

const requests = ref<FriendRequest[]>([])
const requestsLoading = ref(true)

const friendList = ref<Friend[]>([])
const friendsLoading = ref(true)

const actionLoading = ref<Record<number, string>>({})

// ── Data fetching ──────────────────────────────────────────
async function loadRequests() {
  requestsLoading.value = true
  try {
    const res = await friendsApi.getFriendRequests()
    requests.value = res.data ?? []
  } catch {
    requests.value = []
  } finally {
    requestsLoading.value = false
  }
}

async function loadFriends() {
  friendsLoading.value = true
  try {
    const res = await friendsApi.getFriends()
    friendList.value = res.data ?? []
  } catch {
    friendList.value = []
  } finally {
    friendsLoading.value = false
  }
}

onMounted(() => {
  loadRequests()
  loadFriends()
})

// ── Actions ────────────────────────────────────────────────

async function sendRequest() {
  const raw = friendInput.value.trim()
  if (!raw) return
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(raw)) {
    sendError.value = '请输入有效的邮箱'
    return
  }
  sendLoading.value = true
  sendError.value = ''
  sendSuccess.value = ''
  try {
    await friendsApi.sendFriendRequest(raw)
    sendSuccess.value = '好友请求已发送'
    friendInput.value = ''
  } catch (e: any) {
    sendError.value = e?.response?.data?.message || '发送失败，请重试'
  } finally {
    sendLoading.value = false
  }
}

async function accept(id: number) {
  actionLoading.value[id] = 'accept'
  try {
    await friendsApi.acceptRequest(id)
    requests.value = requests.value.filter((r) => r.id !== id)
    await loadFriends()
  } catch (e: any) {
    window.alert(e?.response?.data?.message || '操作失败')
  } finally {
    delete actionLoading.value[id]
  }
}

async function reject(id: number) {
  actionLoading.value[id] = 'reject'
  try {
    await friendsApi.rejectRequest(id)
    requests.value = requests.value.filter((r) => r.id !== id)
  } catch (e: any) {
    window.alert(e?.response?.data?.message || '操作失败')
  } finally {
    delete actionLoading.value[id]
  }
}

async function remove(id: number) {
  if (!window.confirm('确定要删除这个好友吗？')) return
  actionLoading.value[id] = 'delete'
  try {
    await friendsApi.deleteFriend(id)
    friendList.value = friendList.value.filter((f) => f.id !== id)
  } catch (e: any) {
    window.alert(e?.response?.data?.message || '操作失败')
  } finally {
    delete actionLoading.value[id]
  }
}

function displayName(req: FriendRequest): string {
  return req.from_nickname || req.from_email || `#${req.from_user_id}`
}

function friendDisplayName(f: Friend): string {
  return f.friend_nickname || f.friend_email || `#${f.friend_id}`
}
</script>

<template>
  <div v-if="!isAuthenticated" class="friend-panel friend-panel--locked">
    <div class="empty-state">
      <span class="empty-icon">🔒</span>
      <span>请先登录以使用好友功能</span>
    </div>
  </div>

  <div v-else class="friend-panel">
    <!-- ═══ 添加好友 ═══ -->
    <section class="panel-section">
      <h2 class="section-title">
        <span class="title-bracket">【</span>添加好友<span class="title-bracket">】</span>
      </h2>
      <div class="add-friend-row">
        <div class="input-wrapper">
          <span class="input-prefix">&gt;</span>
          <input
            v-model="friendInput"
            type="text"
            class="cyber-input"
            placeholder="输入好友的邮箱"
            @keyup.enter="sendRequest"
          />
        </div>
        <button
          class="cyber-btn cyber-btn--cyan"
          :disabled="sendLoading"
          @click="sendRequest"
        >
          <span v-if="sendLoading" class="btn-loading">▋</span>
          <span v-else>发送请求</span>
        </button>
      </div>
      <p v-if="sendError" class="msg msg--error">⚠ {{ sendError }}</p>
      <p v-if="sendSuccess" class="msg msg--success">✓ {{ sendSuccess }}</p>
    </section>

    <!-- ═══ 好友请求 ═══ -->
    <section class="panel-section">
      <h2 class="section-title">
        <span class="title-bracket">【</span>好友请求<span class="title-bracket">】</span>
        <span v-if="requests.length" class="badge">{{ requests.length }}</span>
      </h2>
      <div v-if="requestsLoading" class="loading-indicator">
        <span class="cursor-blink">▌</span> 加载中...
      </div>
      <div v-else-if="requests.length === 0" class="empty-state">
        <span class="empty-icon">∅</span>
        <span>暂无待处理的好友请求</span>
      </div>
      <ul v-else class="item-list">
        <li v-for="req in requests" :key="req.id" class="item-row">
          <span class="item-name">{{ displayName(req) }}</span>
          <span class="item-meta">{{ req.status }}</span>
          <div class="item-actions">
            <button
              class="cyber-btn cyber-btn--green cyber-btn--sm"
              :disabled="actionLoading[req.id] === 'accept'"
              @click="accept(req.id)"
            >
              {{ actionLoading[req.id] === 'accept' ? '...' : '接受' }}
            </button>
            <button
              class="cyber-btn cyber-btn--red cyber-btn--sm"
              :disabled="actionLoading[req.id] === 'reject'"
              @click="reject(req.id)"
            >
              {{ actionLoading[req.id] === 'reject' ? '...' : '拒绝' }}
            </button>
          </div>
        </li>
      </ul>
    </section>

    <!-- ═══ 好友列表 ═══ -->
    <section class="panel-section">
      <h2 class="section-title">
        <span class="title-bracket">【</span>好友列表<span class="title-bracket">】</span>
        <span v-if="friendList.length" class="badge">{{ friendList.length }}</span>
      </h2>
      <div v-if="friendsLoading" class="loading-indicator">
        <span class="cursor-blink">▌</span> 加载中...
      </div>
      <div v-else-if="friendList.length === 0" class="empty-state">
        <span class="empty-icon">∅</span>
        <span>暂无好友</span>
      </div>
      <ul v-else class="item-list">
        <li v-for="f in friendList" :key="f.id" class="item-row">
          <span class="item-name item-name--friend">{{ friendDisplayName(f) }}</span>
          <button
            class="cyber-btn cyber-btn--pink cyber-btn--sm"
            :disabled="actionLoading[f.id] === 'delete'"
            @click="remove(f.id)"
          >
            {{ actionLoading[f.id] === 'delete' ? '...' : '删除' }}
          </button>
        </li>
      </ul>
    </section>
  </div>
</template>

<style scoped>
/* ═══════════════════════════════════════════════════════════
   FriendPanel — Cyberpunk Retro Terminal Theme
   ═══════════════════════════════════════════════════════════ */

.friend-panel {
  --section-gap: 24px;
  display: flex;
  flex-direction: column;
  gap: var(--section-gap);
  padding: 20px;
  background: rgba(10, 10, 10, 0.92);
  border: 1px solid rgba(0, 255, 255, 0.15);
  border-radius: 4px;
  box-shadow:
    inset 0 0 60px rgba(0, 255, 255, 0.03),
    0 0 20px rgba(0, 255, 255, 0.06);
  backdrop-filter: blur(4px);
  position: relative;
  overflow: hidden;
}

/* Scanline overlay */
.friend-panel::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(0, 0, 0, 0.15) 2px,
    rgba(0, 0, 0, 0.15) 4px
  );
  pointer-events: none;
  z-index: 10;
}

/* ── Section ─────────────────────────────────────────────── */
.panel-section {
  position: relative;
  z-index: 1;
}

.panel-section + .panel-section {
  border-top: 1px solid rgba(0, 255, 255, 0.08);
  padding-top: var(--section-gap);
}

/* ── Section title ───────────────────────────────────────── */
.section-title {
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 700;
  color: var(--neon-cyan);
  text-shadow: 0 0 8px rgba(0, 255, 255, 0.5), 0 0 16px rgba(0, 255, 255, 0.2);
  letter-spacing: 2px;
  margin-bottom: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-bracket {
  color: rgba(0, 255, 255, 0.45);
  text-shadow: none;
}

.badge {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  color: #000;
  background: var(--neon-cyan);
  padding: 1px 7px;
  border-radius: 2px;
  text-shadow: none;
  letter-spacing: 0;
}

/* ── Add friend row ──────────────────────────────────────── */
.add-friend-row {
  display: flex;
  gap: 10px;
}

.input-wrapper {
  flex: 1;
  display: flex;
  align-items: center;
  background: rgba(0, 0, 0, 0.5);
  border: 1px solid rgba(0, 255, 255, 0.2);
  border-radius: 2px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.input-wrapper:focus-within {
  border-color: var(--neon-cyan);
  box-shadow: 0 0 12px rgba(0, 255, 255, 0.2), inset 0 0 12px rgba(0, 255, 255, 0.04);
}

.input-prefix {
  font-family: var(--font-mono);
  color: var(--neon-cyan);
  font-size: 14px;
  padding: 0 8px;
  opacity: 0.6;
}

.cyber-input {
  flex: 1;
  background: none;
  border: none;
  outline: none;
  color: #fff;
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 10px 8px 10px 0;
  caret-color: var(--neon-cyan);
}

.cyber-input::placeholder {
  color: rgba(255, 255, 255, 0.25);
  font-style: italic;
}

/* ── Buttons ─────────────────────────────────────────────── */
.cyber-btn {
  font-family: var(--font-display);
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 1px;
  padding: 10px 18px;
  border: 1px solid;
  border-radius: 2px;
  cursor: pointer;
  background: transparent;
  color: #fff;
  transition: all 0.2s ease;
  white-space: nowrap;
  position: relative;
  overflow: hidden;
}

.cyber-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.cyber-btn--sm {
  padding: 5px 12px;
  font-size: 11px;
  letter-spacing: 0.5px;
}

/* Cyan (send) */
.cyber-btn--cyan {
  border-color: var(--neon-cyan);
  color: var(--neon-cyan);
  text-shadow: 0 0 6px rgba(0, 255, 255, 0.5);
  box-shadow: 0 0 8px rgba(0, 255, 255, 0.1), inset 0 0 8px rgba(0, 255, 255, 0.05);
}

.cyber-btn--cyan:hover:not(:disabled) {
  background: rgba(0, 255, 255, 0.1);
  box-shadow: 0 0 16px rgba(0, 255, 255, 0.25), inset 0 0 12px rgba(0, 255, 255, 0.08);
}

/* Green (accept) */
.cyber-btn--green {
  border-color: var(--neon-green);
  color: var(--neon-green);
  text-shadow: 0 0 6px rgba(0, 255, 0, 0.5);
  box-shadow: 0 0 8px rgba(0, 255, 0, 0.1), inset 0 0 8px rgba(0, 255, 0, 0.05);
}

.cyber-btn--green:hover:not(:disabled) {
  background: rgba(0, 255, 0, 0.08);
  box-shadow: 0 0 16px rgba(0, 255, 0, 0.25), inset 0 0 12px rgba(0, 255, 0, 0.08);
}

/* Red (reject) */
.cyber-btn--red {
  border-color: var(--neon-red);
  color: var(--neon-red);
  text-shadow: 0 0 6px rgba(255, 0, 0, 0.5);
  box-shadow: 0 0 8px rgba(255, 0, 0, 0.1), inset 0 0 8px rgba(255, 0, 0, 0.05);
}

.cyber-btn--red:hover:not(:disabled) {
  background: rgba(255, 0, 0, 0.08);
  box-shadow: 0 0 16px rgba(255, 0, 0, 0.25), inset 0 0 12px rgba(255, 0, 0, 0.08);
}

/* Pink (delete) */
.cyber-btn--pink {
  border-color: var(--neon-pink);
  color: var(--neon-pink);
  text-shadow: 0 0 6px rgba(255, 0, 255, 0.5);
  box-shadow: 0 0 8px rgba(255, 0, 255, 0.1), inset 0 0 8px rgba(255, 0, 255, 0.05);
}

.cyber-btn--pink:hover:not(:disabled) {
  background: rgba(255, 0, 255, 0.08);
  box-shadow: 0 0 16px rgba(255, 0, 255, 0.25), inset 0 0 12px rgba(255, 0, 255, 0.08);
}

/* ── Messages ────────────────────────────────────────────── */
.msg {
  font-family: var(--font-mono);
  font-size: 12px;
  margin-top: 8px;
  padding: 0;
}

.msg--error {
  color: var(--neon-red);
  text-shadow: 0 0 6px rgba(255, 0, 0, 0.4);
}

.msg--success {
  color: var(--neon-green);
  text-shadow: 0 0 6px rgba(0, 255, 0, 0.4);
}

/* ── Loading ─────────────────────────────────────────────── */
.loading-indicator {
  font-family: var(--font-mono);
  font-size: 12px;
  color: rgba(0, 255, 255, 0.5);
  letter-spacing: 1px;
}

.cursor-blink {
  display: inline-block;
  animation: blink 1s step-end infinite;
  color: var(--neon-cyan);
}

/* ── Empty state ─────────────────────────────────────────── */
.empty-state {
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: var(--font-display);
  font-size: 13px;
  color: rgba(255, 255, 255, 0.3);
  padding: 12px 0;
}

.empty-icon {
  font-family: var(--font-mono);
  font-size: 18px;
  color: rgba(0, 255, 255, 0.25);
}

/* ── Item list ───────────────────────────────────────────── */
.item-list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: rgba(0, 255, 255, 0.03);
  border: 1px solid rgba(0, 255, 255, 0.06);
  border-radius: 2px;
  transition: border-color 0.2s, background 0.2s;
}

.item-row:hover {
  background: rgba(0, 255, 255, 0.05);
  border-color: rgba(0, 255, 255, 0.15);
}

.item-name {
  flex: 1;
  font-family: var(--font-mono);
  font-size: 13px;
  color: #ddd;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-name--friend {
  color: var(--neon-green);
  text-shadow: 0 0 4px rgba(0, 255, 0, 0.2);
}

.item-meta {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--neon-yellow);
  text-shadow: 0 0 4px rgba(255, 255, 0, 0.3);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.item-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

/* ── Animations ──────────────────────────────────────────── */
@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
</style>
