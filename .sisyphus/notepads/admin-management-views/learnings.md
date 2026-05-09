# Admin Management Views — Learnings

## Design System Patterns (Cyberpunk Neon)
- **Colors**: `--bg-primary: #0a0a0a` background, neon palette (`--neon-red`, `--neon-green`, `--neon-pink`, `--neon-cyan`, `--neon-yellow`)
- **Fonts**: `--font-display` for titles ("Microsoft YaHei", "SimHei"), `--font-mono` for labels/code ("Courier New")
- **Boxes**: `.neon-box` (green border+glow), `.neon-box--pink`, `.neon-box--red` variants
- **Text glow**: `.glow-text` with layered text-shadows
- **Scanlines**: `.scanlines` overlay with `repeating-linear-gradient` for CRT effect
- **Buttons**: `.action-btn` pattern with color variants and glow on hover
- **Inputs**: `.neon-input` with inset glow + border glow on focus

## API Notes
- `getImages()` params use `page_size` NOT `limit`
- Album update/delete APIs exist in backend but were missing from frontend `albums.ts`
- Backend album routes: POST /api/albums (auth), PUT /api/albums/:id (auth), DELETE /api/albums/:id (auth)

## File Structure
- All admin views under `frontend/src/views/`
- Router already configured with `requiresAdmin` meta guards
- Admin nav: `/admin` → `/admin/images`, `/admin/users`, `/admin/albums`
