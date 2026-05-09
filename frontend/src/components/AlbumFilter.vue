<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Album {
  id: number
  name: string
}

const emit = defineEmits<{
  select: [albumId: number | null]
}>()

const albums = ref<Album[]>([])
const loading = ref(true)
const selectedId = ref<number | null>(null)

async function fetchAlbums() {
  try {
    const res = await fetch('/api/albums')
    if (!res.ok) {
      albums.value = []
      return
    }
    const data = await res.json()
    // Support both { albums: [...] } and [...] response shapes
    albums.value = Array.isArray(data) ? data : (data.albums ?? [])
  } catch {
    // Graceful degradation: show nothing on error
    albums.value = []
  } finally {
    loading.value = false
  }
}

function selectAlbum(id: number | null) {
  selectedId.value = id
  emit('select', id)
}

onMounted(() => {
  fetchAlbums()
})
</script>

<template>
  <div v-if="!loading && albums.length > 0" class="album-filter">
    <button
      class="filter-tag"
      :class="{ 'filter-tag--active': selectedId === null }"
      @click="selectAlbum(null)"
    >
      全部
    </button>
    <button
      v-for="album in albums"
      :key="album.id"
      class="filter-tag"
      :class="{ 'filter-tag--active': selectedId === album.id }"
      @click="selectAlbum(album.id)"
    >
      {{ album.name }}
    </button>
  </div>
</template>

<style scoped>
.album-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 4px 0;
}

.filter-tag {
  padding: 4px 14px;
  font-family: var(--font-mono);
  font-size: 13px;
  color: #888;
  background: transparent;
  border: 2px solid #333;
  border-radius: 4px;
  cursor: pointer;
  transition:
    color 0.2s,
    border-color 0.2s,
    box-shadow 0.2s;
  outline: none;
}

.filter-tag:hover {
  color: var(--neon-green);
  border-color: var(--neon-green);
}

.filter-tag--active {
  color: var(--neon-green);
  border-color: var(--neon-green);
  box-shadow: var(--glow-green);
  background: rgba(0, 255, 0, 0.05);
}

.filter-tag:focus-visible {
  box-shadow: var(--glow-cyan);
  border-color: var(--neon-cyan);
}
</style>
