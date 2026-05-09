<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getImages, searchImages, type ImageData } from '@/api/images'
import ParticleBg from '@/components/ParticleBg.vue'
import LayoutShell from '@/components/LayoutShell.vue'
import TopNotifyBar from '@/components/TopNotifyBar.vue'
import Marquee from '@/components/Marquee.vue'
import RainbowBanner from '@/components/RainbowBanner.vue'
import SearchBar from '@/components/SearchBar.vue'
import AlbumFilter from '@/components/AlbumFilter.vue'
import GalleryGrid from '@/components/GalleryGrid.vue'
import CommentSection from '@/components/CommentSection.vue'
import StatsBar from '@/components/StatsBar.vue'
import VisitorCounter from '@/components/VisitorCounter.vue'
import FloatingButtons from '@/components/FloatingButtons.vue'
import ImagePopup from '@/components/ImagePopup.vue'

// ── Local image type (matches component interface) ──
interface GalleryImage {
  id: number
  title: string
  lsky_url: string
  tags: string[]
  uploaded_by: number
  created_at: string
}

function mapImage(api: ImageData): GalleryImage {
  return {
    id: api.id,
    title: api.title,
    lsky_url: api.url,
    tags: api.tags ?? [],
    uploaded_by: api.user_id,
    created_at: api.created_at,
  }
}

// ── Router & Auth ──
const router = useRouter()
const auth = useAuthStore()

// ── Gallery state ──
const images = ref<GalleryImage[]>([])
const loading = ref(true)
const loadingMore = ref(false)
const hasMore = ref(false)
const currentPage = ref(1)
const currentAlbumId = ref<number | null>(null)
const currentQuery = ref('')
const searchNoResults = ref(false)

const pageSize = 12

// ── Popup state ──
const selectedImage = ref<GalleryImage | null>(null)
const showPopup = ref(false)

// ── Visitor count (placeholder, wired later) ──
const visitorCount = ref(0)

// ── Fetch ──
async function fetchImages(reset = false) {
  if (reset) {
    currentPage.value = 1
    images.value = []
    searchNoResults.value = false
    loading.value = true
  } else {
    loadingMore.value = true
  }

  try {
    const params: Record<string, unknown> = {
      page: currentPage.value,
      limit: pageSize,
    }
    if (currentAlbumId.value) {
      params.album_id = currentAlbumId.value
    }

    let res
    if (currentQuery.value) {
      res = await searchImages(currentQuery.value, {
        page: currentPage.value,
        limit: pageSize,
        ...(currentAlbumId.value ? { album_id: currentAlbumId.value } : {}),
      })
    } else {
      res = await getImages({
        page: currentPage.value,
        limit: pageSize,
        ...(currentAlbumId.value ? { album_id: currentAlbumId.value } : {}),
      })
    }

    const pageData = res.data
    // API returns { list: [...] } for getImages, { images: [...] } for searchImages
    const items = (pageData as any).list || (pageData as any).images || []
    const mapped = items.map(mapImage)

    if (reset) {
      images.value = mapped
    } else {
      images.value.push(...mapped)
    }

    const totalPages = Math.ceil((pageData as any).total / pageSize) || 0
    hasMore.value = (pageData as any).page < totalPages

    if (currentQuery.value && mapped.length === 0 && reset) {
      searchNoResults.value = true
    }
  } catch {
    if (reset) images.value = []
    hasMore.value = false
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

// ── Event handlers ──
function onSearch(query: string) {
  currentQuery.value = query
  currentAlbumId.value = null
  fetchImages(true)
}

function onAlbumSelect(albumId: number | null) {
  currentAlbumId.value = albumId
  currentQuery.value = ''
  fetchImages(true)
}

function onLoadMore() {
  if (loadingMore.value || !hasMore.value) return
  currentPage.value++
  fetchImages(false)
}

function onCardClick(imageId: number) {
  const img = images.value.find((i) => i.id === imageId)
  if (img) {
    selectedImage.value = img
    showPopup.value = true
  }
}

function onPopupClose() {
  showPopup.value = false
}

// ── Lifecycle ──
onMounted(() => {
  fetchImages(true)
})

// ── Computed ──
const showComments = computed(
  () => showPopup.value && selectedImage.value !== null
)

const popupImage = computed(() => selectedImage.value)
</script>

<template>
  <ParticleBg />
  <LayoutShell>
    <!-- ── Top bar ── -->
    <template #top>
      <TopNotifyBar
        online-count="--"
        new-members="--"
        uptime="--"
      />
    </template>

    <!-- ── Left sidebar ── -->
    <template #left-sidebar>
      <Marquee
        text="🔥 欢迎来到群友风采展示站 · 奥力给 · 冲冲冲 🔥"
        color="green"
      />
      <Marquee
        text="💪 群友出征 · 寸草不生 · 天下无敌 💪"
        color="pink"
      />
      <Marquee
        text="⚡ CYBERPUNK 2077 · 霓虹永不熄灭 ⚡"
        color="yellow"
      />
    </template>

    <!-- ── Main area ── -->
    <template #main>
      <RainbowBanner
        title="群友风采"
        subtitle="Official Fan Site"
      />

      <div class="search-row">
        <SearchBar @search="onSearch" />
        <AlbumFilter @select="onAlbumSelect" />
      </div>

      <GalleryGrid
        v-if="!searchNoResults"
        :images="images"
        :loading="loading"
        :has-more="hasMore"
        @load-more="onLoadMore"
        @card-click="onCardClick"
      />

      <div
        v-if="searchNoResults && !loading"
        class="no-results neon-box"
      >
        <span class="no-results-icon">🕵️</span>
        <span>未找到相关图片</span>
      </div>

      <CommentSection
        v-if="showComments"
        :key="selectedImage?.id"
        :image-id="selectedImage!.id"
        :is-logged-in="auth.isLoggedIn"
        :current-user-id="auth.user?.id ?? 0"
        :is-admin="auth.isAdmin"
      />
    </template>

    <!-- ── Right sidebar ── -->
    <template #right-sidebar>
      <StatsBar />
      <div class="visitor-wrap">
        <VisitorCounter :count="visitorCount" />
      </div>
    </template>
  </LayoutShell>

  <!-- ── Floating action buttons ── -->
  <FloatingButtons
    :is-logged-in="auth.isLoggedIn"
    @login="router.push('/login')"
    @register="router.push('/register')"
  />

  <!-- ── Image popup ── -->
  <ImagePopup
    v-if="popupImage"
    :image="popupImage"
    :visible="showPopup"
    @close="onPopupClose"
  />
</template>

<style scoped>
.search-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
  flex-wrap: wrap;
}

.search-row > :first-child {
  flex: 1;
  min-width: 200px;
}

.no-results {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px 20px;
  margin-top: 16px;
  font-family: var(--font-display);
  font-size: 18px;
  color: #555;
}

.no-results-icon {
  font-size: 48px;
  opacity: 0.5;
}

.visitor-wrap {
  margin-top: 16px;
  text-align: center;
}
</style>
