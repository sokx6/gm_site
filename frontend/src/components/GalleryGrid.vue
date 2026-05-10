<script setup lang="ts">
import GalleryCard from './GalleryCard.vue'

interface Image {
  id: number
  title: string
  lsky_url: string
  tags: string[]
  uploaded_by: number
  uploader_name: string
  created_at: string
}

defineProps<{
  images: Image[]
  loading: boolean
  hasMore: boolean
}>()

const emit = defineEmits<{
  'load-more': []
  'card-click': [id: number]
}>()
</script>

<template>
  <div class="gallery-grid">
    <div v-if="images.length === 0 && !loading" class="gallery-empty">
      <span class="empty-icon">📸</span>
      <span>暂无图片，快来上传吧！</span>
    </div>

    <div v-if="loading" class="gallery-loading neon-box">
      <span class="loading-text">加载中...</span>
    </div>

    <GalleryCard
      v-for="image in images"
      :key="image.id"
      :image="image"
      @click="(id: number) => emit('card-click', id)"
    />

    <div v-if="hasMore" class="gallery-load-more-wrap">
      <button class="load-more-btn neon-box" @click="emit('load-more')">
        加载更多
      </button>
    </div>
  </div>
</template>

<style scoped>
.gallery-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
  width: 100%;
}

.gallery-empty {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px 20px;
  font-family: var(--font-display);
  font-size: 18px;
  color: #555;
}

.empty-icon {
  font-size: 48px;
  opacity: 0.5;
}

.gallery-loading {
  grid-column: 1 / -1;
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

.gallery-load-more-wrap {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  padding: 16px 0;
}

.load-more-btn {
  font-family: var(--font-mono);
  font-size: 14px;
  color: var(--neon-cyan);
  background: var(--bg-primary);
  border: 2px double var(--neon-cyan);
  box-shadow: 0 0 8px var(--neon-cyan);
  padding: 8px 32px;
  border-radius: 4px;
  cursor: pointer;
  transition:
    box-shadow 0.25s ease,
    border-color 0.25s ease,
    color 0.25s ease;
}

.load-more-btn:hover {
  border-color: var(--neon-yellow);
  color: var(--neon-yellow);
  box-shadow:
    0 0 16px var(--neon-yellow),
    0 0 32px var(--neon-cyan);
}

@media (max-width: 768px) {
  .gallery-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 10px;
  }
}
</style>
