<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const props = withDefaults(
  defineProps<{
    title?: string
    subtitle?: string
  }>(),
  {
    title: '群友风采',
    subtitle: 'Official Fan Site',
  }
)

const isGlitching = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  timer = setInterval(() => {
    isGlitching.value = true
    setTimeout(() => {
      isGlitching.value = false
    }, 400)
  }, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="rainbow-banner rainbow-border animate-bg">
    <h1 :class="['banner-title', 'glow-text', { 'glitch-active': isGlitching }]">
      ⚡ {{ props.title }} ⚡
    </h1>
    <p class="banner-subtitle">{{ props.subtitle }}</p>
  </div>
</template>

<style scoped>
.rainbow-banner {
  width: 100%;
  padding: 24px 16px;
  text-align: center;
  position: relative;
  z-index: 1;
}

.animate-bg {
  animation: rainbow-bg 6s linear infinite;
}

.banner-title {
  font-size: clamp(20px, 4vw, 36px);
  margin: 0 0 8px;
  line-height: 1.3;
}

.banner-subtitle {
  font-family: var(--font-mono);
  font-size: 14px;
  color: var(--neon-yellow);
  text-transform: uppercase;
  letter-spacing: 4px;
  animation: blink 2s ease-in-out infinite;
}

.glitch-active {
  animation: glitch 0.4s ease;
}
</style>
