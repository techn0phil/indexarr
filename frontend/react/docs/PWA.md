# Progressive Web App (PWA) Implementation

Indexarr is now a fully functional Progressive Web App, enabling installation on desktop and mobile devices with offline-first caching strategies.

## Features Enabled

### 1. **Installation Capability**
- Install on desktop (Windows, macOS, Linux)
- Install on mobile (iOS via Add to Home Screen, Android via browser menu)
- Native app-like experience with standalone display mode
- Custom splash screen and home screen icon
- Custom status bar styling

### 2. **Service Worker & Offline Support**
- Auto-registering service worker with automatic updates
- Pre-caching of static assets (HTML, CSS, JS, images, SVG)
- Runtime caching strategies:
  - **NetworkFirst** for API requests (always fetch latest, fallback to cache)
  - **CacheFirst** for TMDB and TVDB metadata (reduce API quota usage)
  - Network timeout of 3 seconds for fast fallback

### 3. **Web App Manifest**
- Complete manifest configuration with metadata
- Multi-size icon set (192x192 and 512x512 with maskable variants)
- App store screenshots (narrow 540x720 and wide 1280x720)
- Standalone display mode
- Teal theme color matching Indexarr branding (#1D9E75)

### 4. **Apple iOS Support**
- Apple-specific meta tags for PWA capabilities
- Safe-area handling for notched devices
- Custom status bar styling (black-translucent)
- Apple touch icon configuration

## File Structure

### Configuration Files
- **vite.config.ts** — PWA plugin configuration with manifest and Workbox settings
- **index.html** — Meta tags for PWA and Apple PWA capabilities
- **src/main.tsx** — PWA service worker registration

### Generated Files (build output)
- **dist/manifest.webmanifest** — PWA manifest with metadata and icons
- **dist/sw.js** — Service worker
- **dist/workbox-*.js** — Workbox runtime library
- **dist/pwa-*.png** — App icons (192x192, 512x512, maskable variants)
- **dist/screenshot-*.png** — App store screenshots

### Icon Generation
- **scripts/generate-pwa-icons.mjs** — Script to generate PNG icons from canvas
- **public/pwa-*.png** — Source icons (copied to dist on build)

## Installation & Usage

### Development
```bash
# Start dev server with PWA functionality
npm run dev

# Service worker registration works in dev mode (devOptions enabled)
# Open browser DevTools → Application → Service Workers to verify
```

### Production
```bash
# Build with PWA generation
npm run build

# Preview production build
npm run preview
```

### Icon Regeneration
If you need to update PWA icons (e.g., change colors, branding):
```bash
npm run pwa:icons
```

This regenerates all PNG icons and screenshots in `public/` directory.

## Caching Strategies Explained

### Pre-cached Assets
- All static files (`.js`, `.css`, `.html`, `.svg`, `.png`, `.ico`)
- Generated during build via Workbox

### Runtime Caching

**API Requests** (`/api/*`)
- Strategy: NetworkFirst
- Timeout: 3 seconds
- Cache duration: 5 minutes
- Always tries to fetch fresh data; falls back to cache if offline/timeout

**TMDB & TVDB Metadata**
- Strategy: CacheFirst
- Cache duration: 30 days
- Reduces API quota usage by caching metadata indefinitely

## Browser DevTools Verification

### Chrome/Edge DevTools
1. Open DevTools (F12)
2. Go to **Application** tab
3. **Service Workers** section — Verify `sw.js` is registered and active
4. **Manifest** section — Verify manifest.webmanifest is loaded
5. **Cache Storage** — View cached assets and API responses
6. **Storage** — View app data and preferences

### Firefox DevTools
1. Go to **about:debugging**
2. Click **This Firefox**
3. Scroll to **Service Workers** — Verify registration
4. Application tab shows storage and cache data

## Installation Methods

### Desktop (Chrome/Edge)
1. Open Indexarr in Chrome or Edge
2. Click install icon (top-right address bar) or see install prompt
3. "Install Indexarr" dialog appears
4. Confirm installation
5. App launches in standalone window

### Android
1. Open Indexarr in Chrome or Edge
2. Tap menu (⋮) → "Install app"
3. Confirm installation
4. App appears on home screen with custom icon and splash screen

### iOS (Safari)
1. Open Indexarr in Safari
2. Tap Share button (↗)
3. Tap "Add to Home Screen"
4. Confirm
5. App launches in full-screen mode with custom icon

## Lighthouse PWA Audit

Run Chrome DevTools Lighthouse audit to verify PWA compliance:
1. Open DevTools (F12)
2. **Lighthouse** tab
3. Select **Mobile** (recommended)
4. Click **Analyze page load**
5. Verify PWA audit passes all checks

Expected results:
- ✅ All required manifest fields present
- ✅ Service worker registered and responds to requests
- ✅ Page is installable
- ✅ Offline support functional
- ✅ HTTPS/secure context enabled

## API Strategy Rationale

### Why NetworkFirst for `/api/*`?
- Users expect real-time media data
- Offline mode shows cached results (5-min old) if network unavailable
- 3-second timeout prevents hanging on slow connections

### Why CacheFirst for TMDB/TVDB?
- Metadata rarely changes for same media ID
- Reduces API quota usage significantly
- 30-day cache provides good balance between freshness and performance

## Offline Behavior

When offline:
- ✅ Static pages and assets load from cache
- ✅ Previously viewed media details available from cache
- ✅ UI remains responsive
- ❌ Real-time data (current library state) unavailable
- ❌ New media scans cannot be triggered

When network restored:
- ✅ Service worker automatically detects online state
- ✅ Stale API data refreshed on next request
- ✅ No manual page refresh required (automatic updates enabled)

## Future Enhancements

- [ ] Background sync for queued operations
- [ ] Push notifications for scan completion
- [ ] Periodic background sync for library updates
- [ ] Custom offline landing page with helpful messaging
- [ ] Sync data to cloud (optional WebDAV backend)

## TypeScript Declarations

PWA virtual module types are declared in `src/vite-env.d.ts`:
```typescript
declare module 'virtual:pwa-register' {
  export interface RegisterSWOptions {
    immediate?: boolean
    onNeedRefresh?: () => void
    onOfflineReady?: () => void
    onRegistered?: (registration: ServiceWorkerRegistration) => void
    onRegisterError?: (error: Error) => void
  }
  export function registerSW(options?: RegisterSWOptions): (reloadPage?: boolean) => Promise<void>
}
```

## Troubleshooting

### Service Worker Not Registering
- Verify HTTPS or localhost (required for service workers)
- Check browser console for errors
- Clear site data and reload (DevTools → Application → Clear site data)

### Icons Not Appearing
- Verify icon files exist in `dist/` after build
- Check manifest.webmanifest for correct icon paths
- Browser cache may need clearing

### Offline Access Not Working
- Check Workbox cache in DevTools → Application → Cache Storage
- Verify API endpoint patterns match runtime caching rules
- Test in Chrome DevTools offline mode (Network tab)

### Updates Not Applying
- Service worker auto-updates enabled (`registerType: 'autoUpdate'`)
- Hard refresh (Ctrl+Shift+R) forces immediate update
- Browser automatically checks for updates every 24 hours

## Dependencies

- `vite-plugin-pwa` — PWA plugin for Vite
- `canvas` — Icon generation during build (dev dependency)
- `workbox-*` — Auto-included via vite-plugin-pwa (no manual installation needed)

## Configuration Reference

See `vite.config.ts` for complete PWA configuration including:
- Manifest settings
- Icon definitions
- Screenshot definitions
- Workbox caching strategies
- Runtime caching rules
