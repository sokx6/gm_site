<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const nickname = ref('')
const error = ref('')
const success = ref('')
const loading = ref(false)

function validateForm(): boolean {
  if (!email.value) {
    error.value = '请输入邮箱地址'
    return false
  }
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(email.value)) {
    error.value = '邮箱格式不正确'
    return false
  }
  if (!nickname.value.trim()) {
    error.value = '请输入昵称'
    return false
  }
  if (!password.value) {
    error.value = '请输入密码'
    return false
  }
  if (password.value.length < 6) {
    error.value = '密码长度不能少于6位'
    return false
  }
  return true
}

async function handleSubmit() {
  error.value = ''
  success.value = ''
  if (!validateForm()) return

  loading.value = true
  try {
    await authStore.register(email.value, password.value, nickname.value.trim())
    success.value = '注册成功，请等待管理员审核'
    // Clear form
    email.value = ''
    password.value = ''
    nickname.value = ''
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || ''
    if (msg.includes('已被注册') || msg.includes('exists') || msg.includes('exist')) {
      error.value = '该邮箱已被注册'
    } else if (msg) {
      error.value = msg
    } else {
      error.value = '注册失败，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card neon-box">
      <!-- Decorative scanline overlay -->
      <div class="scanlines" aria-hidden="true"></div>

      <h1 class="auth-title glow-text">⚡ 注 册 ⚡</h1>
      <p class="auth-subtitle">加入群友大家庭</p>

      <!-- Success message -->
      <div v-if="success" class="auth-success">
        <span class="success-icon">✓</span>
        <div class="success-content">
          <p class="success-title">{{ success }}</p>
          <router-link to="/login" class="neon-link">
            前往登录 →
          </router-link>
        </div>
      </div>

      <!-- Form (hidden after success) -->
      <form v-if="!success" class="auth-form" @submit.prevent="handleSubmit">
        <!-- Error message -->
        <div v-if="error" class="auth-error">
          <span class="error-icon">⚠</span>
          {{ error }}
        </div>

        <!-- Email -->
        <div class="form-field">
          <label class="field-label" for="reg-email">
            <span class="label-icon">📧</span> 邮箱
          </label>
          <input
            id="reg-email"
            v-model="email"
            type="email"
            class="neon-input"
            placeholder="your@email.com"
            autocomplete="email"
            :disabled="loading"
          />
        </div>

        <!-- Nickname -->
        <div class="form-field">
          <label class="field-label" for="reg-nickname">
            <span class="label-icon">👤</span> 昵称
          </label>
          <input
            id="reg-nickname"
            v-model="nickname"
            type="text"
            class="neon-input"
            placeholder="你的群内昵称"
            autocomplete="nickname"
            :disabled="loading"
          />
        </div>

        <!-- Password -->
        <div class="form-field">
          <label class="field-label" for="reg-password">
            <span class="label-icon">🔑</span> 密码
          </label>
          <input
            id="reg-password"
            v-model="password"
            type="password"
            class="neon-input"
            placeholder="至少6位字符"
            autocomplete="new-password"
            :disabled="loading"
          />
        </div>

        <!-- Submit -->
        <button
          type="submit"
          class="neon-btn neon-btn--submit"
          :disabled="loading"
        >
          <span v-if="loading" class="btn-spinner"></span>
          <span v-else>⚡ 注册</span>
        </button>
      </form>

      <!-- Login link -->
      <p v-if="!success" class="auth-link">
        已有账号？
        <router-link to="/login" class="neon-link">
          去登录
        </router-link>
      </p>
    </div>

    <!-- Ambient glow orbs -->
    <div class="ambient-orb ambient-orb--green" aria-hidden="true"></div>
    <div class="ambient-orb ambient-orb--cyan" aria-hidden="true"></div>
  </div>
</template>

<style scoped>
/* ── Page layout ───────────────────────────────────── */
.auth-page {
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
.auth-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  padding: 40px 32px;
  background: #000;
  border: 3px double var(--neon-green);
  box-shadow:
    var(--glow-green),
    0 0 60px rgba(0, 255, 0, 0.12);
  display: flex;
  flex-direction: column;
  gap: 24px;
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

/* ── Title ────────────────────────────────────────── */
.auth-title {
  font-family: var(--font-display);
  font-size: 32px;
  text-align: center;
  margin: 0;
  color: #fff;
  text-shadow:
    0 0 10px var(--neon-green),
    0 0 20px var(--neon-cyan),
    0 0 40px var(--neon-yellow),
    3px 3px 0 var(--neon-red),
    -3px -3px 0 var(--neon-blue);
  letter-spacing: 8px;
  position: relative;
  z-index: 3;
}

.auth-subtitle {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-green);
  text-align: center;
  text-transform: uppercase;
  letter-spacing: 4px;
  margin: 0;
}

/* ── Form ─────────────────────────────────────────── */
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
  position: relative;
  z-index: 3;
}

/* ── Success ──────────────────────────────────────── */
.auth-success {
  padding: 16px;
  background: rgba(0, 255, 0, 0.08);
  border: 1px solid var(--neon-green);
  box-shadow: 0 0 16px rgba(0, 255, 0, 0.2);
  display: flex;
  align-items: flex-start;
  gap: 12px;
  position: relative;
  z-index: 3;
  animation: popup-bounce 0.5s ease;
}

.success-icon {
  font-size: 24px;
  color: var(--neon-green);
  text-shadow: 0 0 8px var(--neon-green);
  flex-shrink: 0;
  margin-top: 2px;
}

.success-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.success-title {
  font-family: var(--font-display);
  font-size: 16px;
  color: var(--neon-green);
  margin: 0;
  text-shadow: 0 0 8px var(--neon-green);
}

/* ── Error ────────────────────────────────────────── */
.auth-error {
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
  animation: error-shake 0.4s ease;
}

.error-icon {
  font-size: 16px;
  flex-shrink: 0;
}

@keyframes error-shake {
  0%, 100% { transform: translateX(0); }
  20%      { transform: translateX(-6px); }
  40%      { transform: translateX(6px); }
  60%      { transform: translateX(-4px); }
  80%      { transform: translateX(4px); }
}

/* ── Form field ───────────────────────────────────── */
.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-green);
  letter-spacing: 2px;
  text-transform: uppercase;
  text-shadow: 0 0 8px var(--neon-green);
}

.label-icon {
  font-style: normal;
}

/* ── Neon input ───────────────────────────────────── */
.neon-input {
  font-family: var(--font-mono);
  font-size: 15px;
  padding: 12px 14px;
  background: #0a0a0a;
  color: var(--neon-green);
  border: 2px solid var(--neon-green);
  box-shadow:
    inset 0 0 8px rgba(0, 255, 0, 0.15),
    0 0 8px rgba(0, 255, 0, 0.2);
  outline: none;
  transition: border-color 0.3s, box-shadow 0.3s;
}

.neon-input::placeholder {
  color: rgba(0, 255, 0, 0.25);
}

.neon-input:focus {
  border-color: var(--neon-cyan);
  box-shadow:
    inset 0 0 12px rgba(0, 255, 255, 0.2),
    0 0 16px var(--neon-cyan);
}

.neon-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ── Submit button ────────────────────────────────── */
.neon-btn {
  font-family: var(--font-mono);
  font-size: 16px;
  padding: 14px;
  background: #000;
  color: var(--neon-green);
  border: 2px solid var(--neon-green);
  box-shadow: var(--glow-green);
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 4px;
  transition: all 0.25s ease;
  position: relative;
  overflow: hidden;
}

.neon-btn:hover:not(:disabled) {
  background: rgba(0, 255, 0, 0.1);
  box-shadow:
    0 0 20px var(--neon-green),
    0 0 40px rgba(0, 255, 0, 0.3);
  transform: scale(1.02);
}

.neon-btn:active:not(:disabled) {
  transform: scale(0.98);
}

.neon-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.neon-btn--submit {
  margin-top: 4px;
  color: var(--neon-yellow);
  border-color: var(--neon-yellow);
  box-shadow: var(--glow-yellow);
}

.neon-btn--submit:hover:not(:disabled) {
  background: rgba(255, 255, 0, 0.1);
  box-shadow:
    0 0 20px var(--neon-yellow),
    0 0 40px rgba(255, 255, 0, 0.3);
}

/* ── Spinner ──────────────────────────────────────── */
.btn-spinner {
  display: inline-block;
  width: 18px;
  height: 18px;
  border: 2px solid transparent;
  border-top-color: var(--neon-yellow);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ── Login link ───────────────────────────────────── */
.auth-link {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-green);
  text-align: center;
  margin: 0;
  position: relative;
  z-index: 3;
}

.neon-link {
  color: var(--neon-yellow);
  text-decoration: none;
  text-shadow: 0 0 8px var(--neon-yellow);
  transition: text-shadow 0.3s;
}

.neon-link:hover {
  text-shadow:
    0 0 16px var(--neon-yellow),
    0 0 32px var(--neon-yellow);
  text-decoration: underline;
}

/* ── Ambient orbs ─────────────────────────────────── */
.ambient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  pointer-events: none;
}

.ambient-orb--green {
  width: 240px;
  height: 240px;
  background: var(--neon-green);
  opacity: 0.07;
  top: -60px;
  left: -60px;
  animation: orb-float 8s ease-in-out infinite;
}

.ambient-orb--cyan {
  width: 200px;
  height: 200px;
  background: var(--neon-cyan);
  opacity: 0.06;
  bottom: -40px;
  right: -50px;
  animation: orb-float 10s ease-in-out infinite reverse;
}

@keyframes orb-float {
  0%, 100% { transform: translate(0, 0); }
  50%      { transform: translate(20px, -20px); }
}

/* ── Responsive ───────────────────────────────────── */
@media (max-width: 480px) {
  .auth-card {
    padding: 28px 20px;
    gap: 18px;
  }

  .auth-title {
    font-size: 24px;
    letter-spacing: 4px;
  }

  .neon-btn {
    font-size: 14px;
    padding: 12px;
  }
}
</style>
