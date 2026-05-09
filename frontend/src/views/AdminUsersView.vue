<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getPendingUsers, approveUser, rejectUser, type PendingUser } from '@/api/admin'

// ── State ──
const pendingUsers = ref<PendingUser[]>([])
const approvedUsers = ref<PendingUser[]>([])
const loading = ref(true)
const error = ref('')
const actionMsg = ref('')

// ── Fetch ──
async function fetchPending() {
  loading.value = true
  error.value = ''
  try {
    const res = await getPendingUsers()
    pendingUsers.value = res.data
  } catch (e: any) {
    error.value = e?.response?.data?.message || '加载待审核用户失败'
  } finally {
    loading.value = false
  }
}

onMounted(fetchPending)

// ── Actions ──
async function handleApprove(user: PendingUser) {
  actionMsg.value = ''
  try {
    await approveUser(user.id)
    pendingUsers.value = pendingUsers.value.filter(u => u.id !== user.id)
    approvedUsers.value.unshift(user)
    actionMsg.value = `已批准 ${user.nickname || user.email}`
  } catch (e: any) {
    error.value = e?.response?.data?.message || '批准失败'
  }
}

async function handleReject(user: PendingUser) {
  actionMsg.value = ''
  try {
    await rejectUser(user.id)
    pendingUsers.value = pendingUsers.value.filter(u => u.id !== user.id)
    actionMsg.value = `已拒绝 ${user.nickname || user.email}`
  } catch (e: any) {
    error.value = e?.response?.data?.message || '拒绝失败'
  }
}

// ── Format ──
function formatDate(d: string): string {
  return new Date(d).toLocaleString('zh-CN')
}
</script>

<template>
  <div class="admin-sub-page">
    <!-- Header -->
    <div class="page-header">
      <router-link to="/admin" class="neon-link back-link">← 返回管理</router-link>
      <h1 class="page-title glow-text">👥 用户管理</h1>
      <p class="page-subtitle">待审核用户: {{ pendingUsers.length }} 人</p>
    </div>

    <!-- Error -->
    <div v-if="error" class="page-error">
      <span class="error-icon">⚠</span>
      {{ error }}
      <button class="dismiss-btn" @click="error = ''">×</button>
    </div>

    <!-- Success msg -->
    <div v-if="actionMsg" class="page-success">
      <span class="success-icon">✓</span>
      {{ actionMsg }}
      <button class="dismiss-btn dismiss-btn--green" @click="actionMsg = ''">×</button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-box neon-box">
      <span class="loading-text">加载中...</span>
    </div>

    <!-- Pending users list -->
    <template v-else>
      <section class="user-section">
        <h2 class="section-title">
          <span class="section-icon">⏳</span> 待审核用户
        </h2>

        <div v-if="pendingUsers.length === 0" class="empty-state">
          <span class="empty-icon">✅</span>
          <p>暂无待审核用户</p>
        </div>

        <div class="user-cards">
          <div
            v-for="user in pendingUsers"
            :key="user.id"
            class="user-card neon-box--red"
          >
            <div class="scanlines" aria-hidden="true"></div>

            <div class="user-info">
              <div class="user-row">
                <span class="user-label">📧 邮箱</span>
                <span class="user-value">{{ user.email }}</span>
              </div>
              <div class="user-row">
                <span class="user-label">👤 昵称</span>
                <span class="user-value">{{ user.nickname || '未设置' }}</span>
              </div>
              <div class="user-row">
                <span class="user-label">📅 注册时间</span>
                <span class="user-value">{{ formatDate(user.created_at) }}</span>
              </div>
            </div>

            <div class="user-actions">
              <button
                class="action-btn action-btn--approve"
                @click="handleApprove(user)"
              >
                批准
              </button>
              <button
                class="action-btn action-btn--reject"
                @click="handleReject(user)"
              >
                拒绝
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- Recently approved -->
      <section v-if="approvedUsers.length > 0" class="user-section">
        <h2 class="section-title section-title--green">
          <span class="section-icon">✅</span> 已批准用户
        </h2>
        <div class="user-cards">
          <div
            v-for="user in approvedUsers.slice(0, 10)"
            :key="'approved-' + user.id"
            class="user-card user-card--approved"
          >
            <div class="scanlines" aria-hidden="true"></div>
            <div class="user-info">
              <div class="user-row">
                <span class="user-label">📧 邮箱</span>
                <span class="user-value">{{ user.email }}</span>
              </div>
              <div class="user-row">
                <span class="user-label">👤 昵称</span>
                <span class="user-value">{{ user.nickname || '未设置' }}</span>
              </div>
            </div>
            <div class="approved-badge">已批准</div>
          </div>
        </div>
      </section>
    </template>

    <!-- Refresh -->
    <div v-if="!loading" class="refresh-bar">
      <button class="refresh-btn" @click="fetchPending">🔄 刷新列表</button>
    </div>
  </div>
</template>

<style scoped>
/* ── Page layout ───────────────────────────────────── */
.admin-sub-page {
  min-height: 100vh;
  background: var(--bg-primary);
  padding: 24px 20px 60px;
  max-width: 900px;
  margin: 0 auto;
}

/* ── Header ───────────────────────────────────────── */
.page-header {
  margin-bottom: 24px;
}

.back-link {
  display: inline-block;
  margin-bottom: 12px;
}

.page-title {
  font-family: var(--font-display);
  font-size: 28px;
  margin: 0 0 4px;
  color: #fff;
  text-shadow:
    0 0 10px var(--neon-pink),
    0 0 20px var(--neon-cyan),
    3px 3px 0 var(--neon-red),
    -3px -3px 0 var(--neon-blue);
  letter-spacing: 4px;
}

.page-subtitle {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-cyan);
  margin: 0;
}

.neon-link {
  font-family: var(--font-mono);
  font-size: 14px;
  color: var(--neon-cyan);
  text-decoration: none;
  text-shadow: 0 0 8px var(--neon-cyan);
  transition: color 0.3s, text-shadow 0.3s;
}

.neon-link:hover {
  color: var(--neon-yellow);
  text-shadow: 0 0 12px var(--neon-yellow);
}

/* ── Messages ─────────────────────────────────────── */
.page-error {
  padding: 10px 14px;
  background: rgba(255, 0, 0, 0.12);
  border: 1px solid var(--neon-red);
  color: var(--neon-red);
  font-family: var(--font-mono);
  font-size: 13px;
  box-shadow: 0 0 12px rgba(255, 0, 0, 0.25);
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  animation: error-shake 0.4s ease;
}

.page-success {
  padding: 10px 14px;
  background: rgba(0, 255, 0, 0.08);
  border: 1px solid var(--neon-green);
  color: var(--neon-green);
  font-family: var(--font-mono);
  font-size: 13px;
  box-shadow: 0 0 12px rgba(0, 255, 0, 0.2);
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.error-icon, .success-icon { font-size: 16px; flex-shrink: 0; }

.dismiss-btn {
  margin-left: auto;
  background: none;
  border: none;
  color: var(--neon-red);
  font-size: 18px;
  cursor: pointer;
  font-family: var(--font-mono);
}

.dismiss-btn--green {
  color: var(--neon-green);
}

@keyframes error-shake {
  0%, 100% { transform: translateX(0); }
  20%      { transform: translateX(-6px); }
  40%      { transform: translateX(6px); }
  60%      { transform: translateX(-4px); }
  80%      { transform: translateX(4px); }
}

/* ── Loading ──────────────────────────────────────── */
.loading-box {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 20px;
}

.loading-text {
  font-family: var(--font-mono);
  font-size: 14px;
  color: var(--neon-green);
  animation: blink 1.5s ease-in-out infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50%      { opacity: 0; }
}

/* ── Section ──────────────────────────────────────── */
.user-section {
  margin-bottom: 40px;
}

.section-title {
  font-family: var(--font-display);
  font-size: 18px;
  color: var(--neon-pink);
  margin: 0 0 16px;
  text-shadow: 0 0 8px var(--neon-pink);
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-icon {
  font-style: normal;
}

.section-title--green {
  color: var(--neon-green);
  text-shadow: 0 0 8px var(--neon-green);
}

/* ── Empty ────────────────────────────────────────── */
.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: #555;
  font-family: var(--font-mono);
  font-size: 14px;
}

.empty-icon {
  font-size: 32px;
  display: block;
  margin-bottom: 8px;
}

/* ── User cards ───────────────────────────────────── */
.user-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-card {
  position: relative;
  background: #000;
  border: 2px double var(--neon-red);
  box-shadow: var(--glow-red), 0 0 30px rgba(255, 0, 0, 0.06);
  padding: 20px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
  transition: box-shadow 0.3s ease;
}

.user-card:hover {
  box-shadow:
    0 0 20px var(--neon-red),
    0 0 40px rgba(255, 0, 0, 0.15);
}

.user-card--approved {
  border-color: var(--neon-green);
  box-shadow: var(--glow-green), 0 0 30px rgba(0, 255, 0, 0.04);
  opacity: 0.7;
}

.user-card--approved:hover {
  box-shadow:
    0 0 20px var(--neon-green),
    0 0 40px rgba(0, 255, 0, 0.1);
}

.scanlines {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(0, 0, 0, 0.1) 2px,
    rgba(0, 0, 0, 0.1) 4px
  );
  z-index: 2;
}

.user-info {
  position: relative;
  z-index: 3;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.user-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-wrap: wrap;
}

.user-label {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--neon-pink);
  text-transform: uppercase;
  letter-spacing: 1px;
  flex-shrink: 0;
  text-shadow: 0 0 6px var(--neon-pink);
}

.user-value {
  font-family: var(--font-mono);
  font-size: 14px;
  color: #ddd;
  word-break: break-all;
}

/* ── Actions ──────────────────────────────────────── */
.user-actions {
  position: relative;
  z-index: 3;
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.action-btn {
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 8px 20px;
  background: #000;
  cursor: pointer;
  border: 1px solid;
  text-transform: uppercase;
  letter-spacing: 2px;
  transition: all 0.25s ease;
}

.action-btn--approve {
  color: var(--neon-green);
  border-color: var(--neon-green);
  box-shadow: var(--glow-green);
}

.action-btn--approve:hover {
  background: rgba(0, 255, 0, 0.1);
  box-shadow: 0 0 20px var(--neon-green), 0 0 40px rgba(0, 255, 0, 0.2);
}

.action-btn--reject {
  color: var(--neon-red);
  border-color: var(--neon-red);
  box-shadow: var(--glow-red);
}

.action-btn--reject:hover {
  background: rgba(255, 0, 0, 0.1);
  box-shadow: 0 0 20px var(--neon-red), 0 0 40px rgba(255, 0, 0, 0.2);
}

/* ── Approved badge ───────────────────────────────── */
.approved-badge {
  position: relative;
  z-index: 3;
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--neon-green);
  border: 1px solid var(--neon-green);
  padding: 4px 10px;
  text-transform: uppercase;
  letter-spacing: 2px;
  box-shadow: 0 0 6px rgba(0, 255, 0, 0.2);
}

/* ── Refresh ──────────────────────────────────────── */
.refresh-bar {
  display: flex;
  justify-content: center;
  margin-top: 32px;
}

.refresh-btn {
  font-family: var(--font-mono);
  font-size: 14px;
  padding: 10px 28px;
  background: #000;
  color: var(--neon-cyan);
  border: 1px solid var(--neon-cyan);
  box-shadow: 0 0 8px rgba(0, 255, 255, 0.15);
  cursor: pointer;
  transition: all 0.3s ease;
}

.refresh-btn:hover {
  background: rgba(0, 255, 255, 0.1);
  box-shadow:
    0 0 16px var(--neon-cyan),
    0 0 32px rgba(0, 255, 255, 0.2);
}

/* ── Responsive ───────────────────────────────────── */
@media (max-width: 600px) {
  .user-card {
    flex-direction: column;
    align-items: stretch;
    padding: 16px;
  }

  .user-actions {
    justify-content: flex-end;
  }

  .action-btn {
    padding: 8px 14px;
    font-size: 12px;
  }
}
</style>
