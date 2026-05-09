<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    text: string
    color?: 'green' | 'pink' | 'yellow'
  }>(),
  {
    color: 'green',
  }
)

const colorVar = computed(() => {
  const map: Record<string, string> = {
    green: 'var(--neon-green)',
    pink: 'var(--neon-pink)',
    yellow: 'var(--neon-yellow)',
  }
  return map[props.color] ?? map.green
})

const boxClass = computed(() => {
  const map: Record<string, string> = {
    green: 'neon-box',
    pink: 'neon-box neon-box--pink',
    yellow: 'neon-box neon-box--red',
  }
  return map[props.color] ?? map.green
})
</script>

<template>
  <div :class="['marquee-wrap', boxClass]">
    <div class="marquee-track">
      <span class="marquee-text" :style="{ color: colorVar }">
        {{ props.text }}
      </span>
      <span class="marquee-text" :style="{ color: colorVar }" aria-hidden="true">
        {{ props.text }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.marquee-wrap {
  overflow: hidden;
  white-space: nowrap;
  width: 100%;
}

.marquee-track {
  display: inline-flex;
  animation: marquee 12s linear infinite;
}

.marquee-text {
  display: inline-block;
  font-family: var(--font-mono);
  font-size: 14px;
  padding-right: 48px;
  text-shadow: 0 0 6px currentColor;
}
</style>
