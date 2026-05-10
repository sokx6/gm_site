<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getImages, searchImages, type ImageData } from '@/api/images'
import { getAlbums, type AlbumData } from '@/api/albums'
import UploadModal from '@/components/UploadModal.vue'
import ParticleBg from '@/components/ParticleBg.vue'
import LayoutShell from '@/components/LayoutShell.vue'
import TopNotifyBar from '@/components/TopNotifyBar.vue'
import Marquee from '@/components/Marquee.vue'
import RainbowBanner from '@/components/RainbowBanner.vue'
import SearchBar from '@/components/SearchBar.vue'
import AlbumFilter from '@/components/AlbumFilter.vue'
import GalleryGrid from '@/components/GalleryGrid.vue'
import StatsBar from '@/components/StatsBar.vue'
import VisitorCounter from '@/components/VisitorCounter.vue'
import FriendPanel from '@/components/FriendPanel.vue'
import FloatingButtons from '@/components/FloatingButtons.vue'
import ImagePopup from '@/components/ImagePopup.vue'
import { onlineCount as topOnline, totalVisitors as visitorCount, newMembers as topNewMembers, uptimeDays as topUptime, connect, disconnect } from '@/composables/useStatsWebSocket'

// ── Local image type (matches component interface) ──
interface GalleryImage {
  id: number
  title: string
  lsky_url: string
  tags: string[]
  uploaded_by: number
  uploader_name: string
  privacy?: string
  created_at: string
}

function mapImage(api: ImageData): GalleryImage {
  return {
    id: api.id,
    title: api.title,
    lsky_url: api.lsky_url || api.url || '',
    tags: api.tags ?? [],
    uploaded_by: api.uploaded_by,
    uploader_name: api.uploader_name || '',
    privacy: api.privacy,
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
const openPopups = ref<GalleryImage[]>([])

// ── Upload state ──
const showUploadModal = ref(false)
const showFriendPanel = ref(false)
const albumList = ref<AlbumData[]>([])

// ── Site name ──
const siteName = ref('群友风采')

async function fetchSiteName() {
  try {
    const res = await fetch('/api/site')
    const data = await res.json()
    if (data.data?.name) siteName.value = data.data.name
  } catch {
    // fallback to default
  }
}



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
    openPopups.value.push(img)
  }
}

function onPopupClose(imageId: number) {
  openPopups.value = openPopups.value.filter((i) => i.id !== imageId)
}

function onUploadClick() {
  showUploadModal.value = true
}

function onUploadModalClose() {
  showUploadModal.value = false
}

function onUploaded() {
  fetchImages(true)
  showUploadModal.value = false
}

async function fetchAlbums() {
  try {
    const res = await getAlbums()
    albumList.value = res.data ?? []
  } catch {
    albumList.value = []
  }
}

// ── Hitokoto quotes ──
const hitokotoQuotes = ref<string[]>([])

async function fetchHitokoto() {
  try {
    const res = await fetch(`https://v1.hitokoto.cn/?c=a&c=b&c=d&_t=${Date.now()}${Math.random()}`)
    const data = await res.json()
    return data.hitokoto + ' —— ' + (data.from || '佚名')
  } catch {
    return null
  }
}

const marqueeCount = ref(3)

function calcMarqueeCount() {
  marqueeCount.value = Math.max(3, Math.floor(window.innerHeight / 80))
}

async function refreshHitokotoQuotes() {
  calcMarqueeCount()
  const fetches = Array.from({ length: marqueeCount.value }, () => fetchHitokoto())
  const results = await Promise.all(fetches)
  const valid = results.filter((q): q is string => q !== null)
  if (valid.length >= 3) {
    hitokotoQuotes.value = valid
  } else {
    hitokotoQuotes.value = Array.from({ length: marqueeCount.value }, () => '加载中...')
  }
}

// ── Lifecycle ──
let hitokotoTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  fetchSiteName()
  fetchImages(true)
  fetchAlbums()
  connect()
  calcMarqueeCount()
  refreshHitokotoQuotes()
  hitokotoTimer = setInterval(refreshHitokotoQuotes, 30000)
  window.addEventListener('resize', calcMarqueeCount)
})

onUnmounted(() => {
  disconnect()
  if (hitokotoTimer) clearInterval(hitokotoTimer)
  window.removeEventListener('resize', calcMarqueeCount)
})

// ── Marquee color helper ──
const marqueeColors = ['green', 'pink', 'yellow'] as const
function marqueeColor(i: number): 'green' | 'pink' | 'yellow' {
  return marqueeColors[i % 3]
}

// ── Computed ──
</script>

<template>
  <ParticleBg />
  <LayoutShell>
    <!-- ── Top bar ── -->
    <template #top>
      <TopNotifyBar
        :online-count="topOnline"
        :new-members="topNewMembers"
        :uptime="topUptime"
      />
    </template>

    <!-- ── Left sidebar ── -->
    <template #left-sidebar>
      <Marquee
        v-for="(quote, i) in hitokotoQuotes"
        :key="i"
        :text="quote"
        :color="marqueeColor(i)"
      />
    </template>

    <!-- ── Main area ── -->
    <template #main>
      <RainbowBanner
        :title="siteName"
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

    </template>

    <!-- ── Right sidebar ── -->
    <template #right-sidebar>
      <StatsBar />
      <div class="visitor-wrap">
        <VisitorCounter :count="visitorCount" />
      </div>
      <FriendPanel v-if="showFriendPanel && auth.isLoggedIn" />
    </template>
  </LayoutShell>

  <!-- ── Floating action buttons ── -->
  <FloatingButtons
    :is-logged-in="auth.isLoggedIn"
    :is-admin="auth.isAdmin"
    @login="router.push('/login')"
    @register="router.push('/register')"
    @upload="onUploadClick"
    @admin="router.push('/admin')"
    @friend="showFriendPanel = !showFriendPanel"
  />

  <UploadModal
    :visible="showUploadModal"
    :albums="albumList"
    @close="onUploadModalClose"
    @uploaded="onUploaded"
  />

  <!-- ── Image popup ── -->
  <ImagePopup
    v-for="img in openPopups"
    :key="img.id"
    :image="img"
    :visible="true"
    :is-logged-in="auth.isLoggedIn"
    :current-user-id="auth.user?.id ?? 0"
    :is-admin="auth.isAdmin"
    @close="onPopupClose(img.id)"
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
