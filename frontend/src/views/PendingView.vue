<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const checking = ref(false)

async function handleCheckStatus() {
  checking.value = true
  try {
    await authStore.tryRestoreSession()
    // If restored and not pending, redirect to home
    if (authStore.isLoggedIn && !authStore.isPending) {
      router.push('/')
      return
    }
    // If still pending, just let the view remain
  } catch {
    // Ignore errors, stay on pending page
  } finally {
    checking.value = false
  }
}
</script>

<template>
  <div class="pending-page">
    <div class="pending-card neon-box--red">
      <!-- Scanline overlay -->
      <div class="scanlines" aria-hidden="true"></div>

      <!-- Pulsing dot indicator -->
      <div class="pulse-container" aria-hidden="true">
        <div class="pulse-ring pulse-ring--outer"></div>
        <div class="pulse-ring pulse-ring--inner"></div>
        <div class="pulse-dot"></div>
      </div>

      <h1 class="pending-title glow-text">⚡ 审核中 ⚡</h1>

      <p class="pending-subtitle">
        您的账号正在审核中
      </p>

      <p class="pending-desc">
        请耐心等待管理员审核，审核通过后即可登录
      </p>

      <!-- Info box -->
      <div class="info-box">
        <div class="info-row">
          <span class="info-label">📋 状态</span>
          <span class="info-value info-value--pending">待审核</span>
        </div>
        <div class="info-row">
          <span class="info-label">📧 邮箱</span>
          <span class="info-value">{{ authStore.user?.email || '—' }}</span>
        </div>
      </div>

      <div class="pending-actions">
        <button
          class="neon-btn neon-btn--check"
          :disabled="checking"
          @click="handleCheckStatus"
        >
          <span v-if="checking" class="btn-spinner"></span>
          <span v-else>🔄 重新检查状态</span>
        </button>

        <router-link to="/" class="neon-link neon-link--home">
          返回首页
        </router-link>
      </div>
    </div>

    <!-- Ambient orbs -->
    <div class="ambient-orb ambient-orb--red" aria-hidden="true"></div>
    <div class="ambient-orb ambient-orb--orange" aria-hidden="true"></div>
  </div>
</template>

<style scoped>
/* ── Page layout ───────────────────────────────────── */
.pending-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  position: relative;
  overflow: hidden;
  padding: 20px;
}

/* ── Card ──────────────────────────────────────────── */
.pending-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 440px;
  padding: 48px 36px 40px;
  background: #000;
  border: 3px double var(--neon-orange);
  box-shadow:
    0 0 15px var(--neon-orange),
    0 0 60px rgba(255, 102, 0, 0.1);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  text-align: center;
}

/* ── Scanline overlay ─────────────────────────────── */
.scanlines {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(0, 0, 0, 0.15) 2px,
    rgba(0, 0, 0, 0.15) 4px
  );
  z-index: 2;
}

/* ── Pulse animation container ────────────────────── */
.pulse-container {
  position: relative;
  width: 80px;
  height: 80px;
  margin-bottom: 4px;
  z-index: 3;
}

.pulse-dot {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 18px;
  height: 18px;
  background: var(--neon-orange);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  box-shadow:
    0 0 12px var(--neon-orange),
    0 0 28px rgba(255, 102, 0, 0.6),
    0 0 48px rgba(255, 102, 0, 0.3);
  animation: pulse-dot 1.6s ease-in-out infinite;
}

.pulse-ring {
  position: absolute;
  top: 50%;
  left: 50%;
  border-radius: 50%;
  border: 2px solid var(--neon-orange);
  transform: translate(-50%, -50%);
  animation: pulse-expand 2s ease-out infinite;
}

.pulse-ring--outer {
  width: 100%;
  height: 100%;
  animation-delay: 0s;
}

.pulse-ring--inner {
  width: 70%;
  height: 70%;
  animation-delay: 0.4s;
}

@keyframes pulse-dot {
  0%, 100% {
    transform: translate(-50%, -50%) scale(1);
    box-shadow:
      0 0 12px var(--neon-orange),
      0 0 28px rgba(255, 102, 0, 0.6);
  }
  50% {
    transform: translate(-50%, -50%) scale(1.4);
    box-shadow:
      0 0 20px var(--neon-orange),
      0 0 48px rgba(255, 102, 0, 0.8);
  }
}

@keyframes pulse-expand {
  0% {
    width: 10%;
    height: 10%;
    opacity: 0.9;
    border-width: 2px;
  }
  100% {
    width: 100%;
    height: 100%;
    opacity: 0;
    border-width: 1px;
  }
}

/* ── Title ────────────────────────────────────────── */
.pending-title {
  font-family: var(--font-display);
  font-size: 32px;
  margin: 0;
  color: #fff;
  text-shadow:
    0 0 10px var(--neon-orange),
    0 0 20px var(--neon-yellow),
    0 0 40px rgba(255, 102, 0, 0.6),
    3px 3px 0 var(--neon-red),
    -3px -3px 0 var(--neon-blue);
  letter-spacing: 8px;
  position: relative;
  z-index: 3;
}

.pending-subtitle {
  font-family: var(--font-display);
  font-size: 20px;
  color: var(--neon-orange);
  text-shadow: 0 0 12px var(--neon-orange);
  margin: 0;
  position: relative;
  z-index: 3;
}

.pending-desc {
  font-family: var(--font-mono);
  font-size: 13px;
  color: #888;
  line-height: 1.7;
  margin: 0;
  position: relative;
  z-index: 3;
}

/* ── Info box ─────────────────────────────────────── */
.info-box {
  width: 100%;
  padding: 14px 16px;
  background: rgba(255, 102, 0, 0.05);
  border: 1px solid rgba(255, 102, 0, 0.25);
  display: flex;
  flex-direction: column;
  gap: 10px;
  position: relative;
  z-index: 3;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-label {
  font-family: var(--font-mono);
  font-size: 12px;
  color: #888;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.info-value {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-cyan);
  text-shadow: 0 0 6px var(--neon-cyan);
  word-break: break-all;
}

.info-value--pending {
  color: var(--neon-orange);
  text-shadow: 0 0 8px var(--neon-orange);
  animation: blink 2s ease-in-out infinite;
}

/* ── Actions ──────────────────────────────────────── */
.pending-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  position: relative;
  z-index: 3;
  margin-top: 4px;
}

.neon-btn {
  font-family: var(--font-mono);
  font-size: 14px;
  padding: 12px 24px;
  background: #000;
  color: var(--neon-cyan);
  border: 2px solid var(--neon-cyan);
  box-shadow: var(--glow-cyan);
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 3px;
  transition: all 0.25s ease;
  position: relative;
  overflow: hidden;
}

.neon-btn:hover:not(:disabled) {
  background: rgba(0, 255, 255, 0.1);
  box-shadow:
    0 0 20px var(--neon-cyan),
    0 0 40px rgba(0, 255, 255, 0.3);
  transform: scale(1.03);
}

.neon-btn:active:not(:disabled) {
  transform: scale(0.97);
}

.neon-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ── Spinner ──────────────────────────────────────── */
.btn-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid transparent;
  border-top-color: var(--neon-cyan);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ── Links ────────────────────────────────────────── */
.neon-link--home {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-green);
  text-decoration: none;
  text-shadow: 0 0 8px var(--neon-green);
  transition: text-shadow 0.3s;
}

.neon-link--home:hover {
  text-shadow:
    0 0 16px var(--neon-green),
    0 0 32px var(--neon-green);
  text-decoration: underline;
}

/* ── Ambient orbs ─────────────────────────────────── */
.ambient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  pointer-events: none;
}

.ambient-orb--red {
  width: 280px;
  height: 280px;
  background: var(--neon-orange);
  opacity: 0.06;
  top: -80px;
  right: -80px;
  animation: orb-float 8s ease-in-out infinite;
}

.ambient-orb--orange {
  width: 220px;
  height: 220px;
  background: var(--neon-yellow);
  opacity: 0.05;
  bottom: -60px;
  left: -60px;
  animation: orb-float 10s ease-in-out infinite reverse;
}

@keyframes orb-float {
  0%, 100% { transform: translate(0, 0); }
  50%      { transform: translate(20px, -20px); }
}

/* ── Responsive ───────────────────────────────────── */
@media (max-width: 480px) {
  .pending-card {
    padding: 36px 24px 32px;
    gap: 16px;
  }

  .pending-title {
    font-size: 24px;
    letter-spacing: 4px;
  }

  .pending-subtitle {
    font-size: 17px;
  }

  .neon-btn {
    font-size: 13px;
    padding: 10px 20px;
  }
}
</style>
