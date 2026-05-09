<script setup lang="ts">
</script>

<template>
  <div class="layout-shell">
    <!-- Top notification bar area -->
    <div class="layout-top">
      <slot name="top" />
    </div>

    <!-- Three-column grid -->
    <div class="layout-grid">
      <aside class="layout-left neon-box">
        <slot name="left-sidebar" />
      </aside>
      <main class="layout-main">
        <slot name="main" />
      </main>
      <aside class="layout-right neon-box">
        <slot name="right-sidebar" />
      </aside>
    </div>
  </div>
</template>

<style scoped>
.layout-shell {
  min-height: 100vh;
  background: var(--bg-primary);
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
}

.layout-top {
  position: relative;
  z-index: 2;
}

.layout-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 220px 1fr 260px;
  gap: 12px;
  padding: 12px;
}

.layout-left,
.layout-right {
  min-height: 300px;
  padding: 12px;
  overflow-y: auto;
}

.layout-main {
  min-height: 400px;
  padding: 12px;
}

/* ── Responsive: single column at 768px ──── */
@media (max-width: 768px) {
  .layout-grid {
    grid-template-columns: 1fr;
    gap: 8px;
    padding: 8px;
  }

  .layout-left,
  .layout-right {
    min-height: auto;
    max-height: 200px;
  }
}
</style>
