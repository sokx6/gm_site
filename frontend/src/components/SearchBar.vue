<script setup lang="ts">
import { ref, onUnmounted } from 'vue'

const emit = defineEmits<{
  search: [keyword: string]
}>()

const keyword = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function emitSearch() {
  emit('search', keyword.value.trim())
}

function onInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    emitSearch()
  }, 300)
}

function onEnter() {
  if (debounceTimer) clearTimeout(debounceTimer)
  emitSearch()
}

function onClear() {
  keyword.value = ''
  if (debounceTimer) clearTimeout(debounceTimer)
  emitSearch()
}

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<template>
  <div class="search-bar neon-box">
    <input
      v-model="keyword"
      type="text"
      class="search-input"
      placeholder="搜索图片…"
      @input="onInput"
      @keyup.enter="onEnter"
    />
    <button
      v-if="keyword"
      class="search-clear"
      @click="onClear"
      aria-label="清除"
    >
      ×
    </button>
    <button
      class="search-btn"
      @click="onEnter"
      aria-label="搜索"
    >
      🔍
    </button>
  </div>
</template>

<style scoped>
.search-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: #fff;
  font-family: var(--font-mono);
  font-size: 14px;
  padding: 6px 4px;
}

.search-input::placeholder {
  color: #555;
}

.search-clear {
  background: none;
  border: none;
  color: var(--neon-red);
  font-size: 20px;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
  font-family: var(--font-mono);
}

.search-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 4px;
  line-height: 1;
  transition: transform 0.15s;
}

.search-btn:hover {
  transform: scale(1.2);
}
</style>
