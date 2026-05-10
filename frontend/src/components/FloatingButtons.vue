<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    isLoggedIn?: boolean
    isAdmin?: boolean
  }>(),
  {
    isLoggedIn: false,
    isAdmin: false,
  }
)

const emit = defineEmits<{
  login: []
  register: []
  upload: []
  admin: []
  friend: []
}>()

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
</script>

<template>
  <div class="floating-buttons">
    <button
      class="float-btn float-btn--top"
      @click="scrollToTop"
      aria-label="回到顶部"
    >
      TOP
    </button>

    <button
      v-if="props.isLoggedIn"
      class="float-btn float-btn--upload"
      @click="emit('upload')"
    >
       上传
    </button>
    <button
      v-if="props.isLoggedIn"
      class="float-btn float-btn--friend"
      @click="emit('friend')"
    >
       好友
    </button>
    <button
      v-else
      class="float-btn float-btn--login"
      @click="emit('login')"
    >
       登录
    </button>
    <button
      v-if="!props.isLoggedIn"
      class="float-btn float-btn--register"
      @click="emit('register')"
    >
       注册
    </button>

    <button
      v-if="props.isAdmin"
      class="float-btn float-btn--admin"
      @click="emit('admin')"
    >
      管理
    </button>
  </div>
</template>

<style scoped>
.floating-buttons {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.float-btn {
  font-family: var(--font-mono);
  font-size: 12px;
  padding: 8px 12px;
  background: #000;
  color: var(--neon-green);
  border: 2px solid var(--neon-green);
  box-shadow: var(--glow-green);
  cursor: pointer;
  transition: background 0.2s, transform 0.15s;
  animation: float-blink 2s ease-in-out infinite;
  white-space: nowrap;
}

.float-btn:hover {
  background: rgba(0, 255, 0, 0.15);
  transform: scale(1.1);
}

.float-btn--top {
  color: var(--neon-cyan);
  border-color: var(--neon-cyan);
  box-shadow: var(--glow-cyan);
  animation: float-blink-cyan 2s ease-in-out infinite;
}

.float-btn--top:hover {
  background: rgba(0, 255, 255, 0.1);
}

.float-btn--login {
  color: var(--neon-yellow);
  border-color: var(--neon-yellow);
  box-shadow: var(--glow-yellow);
  animation-delay: 0.5s;
}

.float-btn--login:hover {
  background: rgba(255, 255, 0, 0.1);
}

.float-btn--register {
  color: var(--neon-pink);
  border-color: var(--neon-pink);
  box-shadow: var(--glow-pink);
  animation-delay: 1s;
}

.float-btn--register:hover {
  background: rgba(255, 0, 255, 0.1);
}

.float-btn--upload {
  color: var(--neon-yellow);
  border-color: var(--neon-yellow);
  box-shadow: var(--glow-yellow);
}

.float-btn--upload:hover {
  background: rgba(255, 255, 0, 0.1);
}

.float-btn--friend {
  color: var(--neon-cyan);
  border-color: var(--neon-cyan);
  box-shadow: var(--glow-cyan);
}

.float-btn--friend:hover {
  background: rgba(0, 255, 255, 0.1);
}

.float-btn--admin {
  color: var(--neon-red);
  border-color: var(--neon-red);
  box-shadow: var(--glow-red);
}

.float-btn--admin:hover {
  background: rgba(255, 0, 0, 0.1);
}

@keyframes float-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

@keyframes float-blink-cyan {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}
</style>
