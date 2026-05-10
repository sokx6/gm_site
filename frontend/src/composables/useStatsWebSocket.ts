// Shared WebSocket composable — singleton pattern
// All consumers share one connection via ref-counting.
import { ref } from 'vue'

// ── Reactive state ──
const onlineCount = ref<string>('--')
const totalVisitors = ref<number>(0)
const newMembers = ref<string>('--')
const uptimeDays = ref<string>('--')

// ── Module-internal singleton state ──
let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let destroyed = false
let refCount = 0

function createConnection() {
  if (destroyed) return

  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = import.meta.env.VITE_WS_URL
    ? `${import.meta.env.VITE_WS_URL}/api/ws`
    : `${protocol}//${location.host}/api/ws`

  ws = new WebSocket(url)

  ws.onmessage = (event: MessageEvent) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'stats') {
        // Support both msg.data and flat msg shapes
        const d = msg.data || msg
        if (d.online !== undefined) onlineCount.value = String(d.online)
        if (d.visitors !== undefined) totalVisitors.value = Number(d.visitors)
        if (d.newMembers !== undefined) newMembers.value = String(d.newMembers)
        if (d.safeDays !== undefined || d.uptimeDays !== undefined) {
          uptimeDays.value = String(d.safeDays ?? d.uptimeDays)
        }
      }
    } catch {
      // Ignore invalid JSON
    }
  }

  ws.onclose = () => {
    if (!destroyed) {
      reconnectTimer = setTimeout(createConnection, 3000)
    }
  }

  ws.onerror = () => {
    ws?.close()
  }
}

/** Connect to the stats WebSocket. Safe to call from multiple components — only one connection is created. */
export function connect() {
  refCount++
  if (refCount > 1) return // already connected

  destroyed = false
  createConnection()
}

/** Disconnect from the stats WebSocket. Only closes when the last consumer disconnects. */
export function disconnect() {
  refCount--
  if (refCount > 0) return // other consumers still need the connection

  destroyed = true
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (ws) {
    ws.onclose = null
    ws.onerror = null
    ws.close()
    ws = null
  }
}

export { onlineCount, totalVisitors, newMembers, uptimeDays }
