<script setup lang="ts">
import { computed } from 'vue'

const emojis = ['🔥', '💪', '👑', '💩', '🐂', '🍺', '⚡', '💀', '🤡', '🎯']

const particles = computed(() =>
  Array.from({ length: 40 }, (_, i) => ({
    id: i,
    emoji: emojis[i % emojis.length],
    left: `${Math.random() * 100}%`,
    duration: `${4 + Math.random() * 10}s`,
    delay: `${Math.random() * 8}s`,
    fontSize: `${14 + Math.random() * 18}px`,
  }))
)

function particleStyle(p: (typeof particles.value)[number]) {
  return {
    '--fall-duration': p.duration,
    '--fall-delay': p.delay,
    left: p.left,
    fontSize: p.fontSize,
  }
}
</script>

<template>
  <div class="particle-bg" aria-hidden="true">
    <span
      v-for="p in particles"
      :key="p.id"
      class="particle"
      :style="particleStyle(p)"
    >
      {{ p.emoji }}
    </span>
  </div>
</template>

<style scoped>
.particle-bg {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
}

.particle {
  position: absolute;
  top: -100px;
  animation: fall var(--fall-duration, 8s) var(--fall-delay, 0s) linear infinite;
  opacity: 0.8;
  will-change: transform;
}

@keyframes fall {
  0% {
    transform: translateY(-100px) rotate(0deg);
    opacity: 0.8;
  }
  80% {
    opacity: 0.4;
  }
  100% {
    transform: translateY(105vh) rotate(720deg);
    opacity: 0;
  }
}
</style>
