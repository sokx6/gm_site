## F3 Manual QA — Bugs Found & Fixed

### CRITICAL: App.vue missing router-view
- **Symptom**: Page showed static placeholder "🎯 主要内容区域" instead of HomeView
- **Root cause**: App.vue had hardcoded LayoutShell content, no `<router-view />`
- **Fix**: Replaced App.vue with minimal `<router-view />`

### MAJOR: Double /api prefix in API client
- **Symptom**: All API calls went to /api/api/... returning 401
- **Root cause**: baseURL "/api" + path "/api/..." = double prefix
- **Fix**: Changed baseURL from "/api" to "/"

### MAJOR: WebSocket URL mismatch
- **Symptom**: WS connection to /ws returned 404
- **Root cause**: Backend endpoint is /api/ws, not /ws
- **Fix**: Updated StatsBar.vue to /api/ws, vite proxy target to ws://localhost:1323

### MEDIUM: API response format mismatch
- **Symptom**: HomeView errored silently (catch block), gallery always empty
- **Root cause**: Backend returns {list:[], total, page, limit} but frontend expected {data:[], total_pages, page, page_size} (double data wrapping)
- **Fix**: Updated HomeView to use (pageData.list || pageData.images), calculated totalPages from total/limit

### QA Test Results
- Guest browsing: Render ✓, Cyberpunk styling ✓, Empty gallery state ✓
- Search: Type/search ✓, No-results message ✓, Clear/reset ✓
- Register: Form fill ✓, Submit ✓, "等待审核" success message ✓
- Login: Pending blocked ✓, Approved login ✓, Redirect to home ✓, Logged-in state (上传/TOP buttons) ✓
- Image popup: NOT TESTED — no images in database (requires Lsky config)
- WebSocket: StatsBar shows "--" (backend StatsService works but frontend placeholder values)
