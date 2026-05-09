<script setup lang="ts">
import { ref } from 'vue'

interface Image {
  id: number
  title: string
  lsky_url: string
  tags: string[]
  uploaded_by: number
  created_at: string
}

const props = defineProps<{ image: Image }>()
const emit = defineEmits<{ click: [id: number] }>()

const imgError = ref(false)

function formatDate(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    })
  } catch {
    return iso
  }
}

const tagColors = [
  'var(--neon-pink)',
  'var(--neon-cyan)',
  'var(--neon-green)',
  'var(--neon-yellow)',
  'var(--neon-orange)',
  'var(--neon-purple)',
]

function tagColor(idx: number): string {
  return tagColors[idx % tagColors.length]
}
</script>

<template>
  <div class="gallery-card" @click="emit('click', props.image.id)">
    <div class="card-img-wrap">
      <img
        v-if="!imgError"
        :src="props.image.lsky_url"
        :alt="props.image.title"
        class="card-img"
        loading="lazy"
        @error="imgError = true"
      />
      <div v-else class="card-img-fallback">加载失败</div>
    </div>

    <div class="card-body">
      <div class="card-label">
        <span class="card-title">{{ props.image.title }}</span>
        <span
          v-for="(tag, i) in props.image.tags"
          :key="tag"
          class="card-tag"
          :style="{ color: tagColor(i), borderColor: tagColor(i) }"
        >
          {{ tag }}
        </span>
      </div>
      <div class="card-time">
        {{ formatDate(props.image.created_at) }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.gallery-card {
  background: var(--bg-primary);
  border: 2px solid var(--neon-cyan);
  border-radius: 6px;
  box-shadow:
    0 0 8px var(--neon-cyan),
    inset 0 0 8px rgba(0, 255, 255, 0.05);
  overflow: hidden;
  cursor: pointer;
  transition:
    transform 0.25s ease,
    box-shadow 0.25s ease,
    border-color 0.25s ease;
}

.gallery-card:hover {
  transform: scale(1.03);
  border-color: var(--neon-pink);
  box-shadow:
    0 0 20px var(--neon-pink),
    0 0 40px var(--neon-cyan),
    inset 0 0 12px rgba(255, 0, 255, 0.08);
}

.card-img-wrap {
  width: 100%;
  aspect-ratio: 4 / 3;
  overflow: hidden;
  background: #111;
}

.card-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.35s ease;
}

.gallery-card:hover .card-img {
  transform: scale(1.06);
}

.card-img-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-mono);
  font-size: 14px;
  color: var(--neon-red);
  background: #1a1a1a;
}

.card-body {
  padding: 10px 12px;
}

.card-label {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.card-title {
  font-family: var(--font-display);
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  text-shadow: 0 0 6px rgba(255, 255, 255, 0.3);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.card-tag {
  font-family: var(--font-mono);
  font-size: 10px;
  padding: 1px 6px;
  border: 1px solid;
  border-radius: 3px;
  white-space: nowrap;
  opacity: 0.85;
}

.card-time {
  font-family: var(--font-mono);
  font-size: 11px;
  color: #666;
}
</style>
