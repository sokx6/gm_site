<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    count?: number
  }>(),
  {
    count: 0,
  }
)

const digits = computed(() => String(props.count).split(''))
</script>

<template>
  <span class="visitor-counter">
    你是第
    <span class="digit-group">
      <span
        v-for="(d, i) in digits"
        :key="i"
        class="digit"
      >{{ d }}</span>
    </span>
    位访客
  </span>
</template>

<style scoped>
.visitor-counter {
  font-family: var(--font-display);
  font-size: 14px;
  color: #888;
}

.digit-group {
  display: inline-flex;
  gap: 2px;
  margin: 0 2px;
}

.digit {
  display: inline-flex;
  font-family: var(--font-mono);
  font-size: 16px;
  color: var(--neon-green);
  text-shadow: 0 0 8px var(--neon-green), 0 0 16px var(--neon-green);
  background: rgba(0, 255, 0, 0.05);
  border: 1px solid rgba(0, 255, 0, 0.3);
  padding: 1px 5px;
  min-width: 20px;
  justify-content: center;
}
</style>
