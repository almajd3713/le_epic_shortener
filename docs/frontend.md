# Frontend

The frontend is a React SPA that lets users shorten URLs, view previously shortened links, and copy short URLs to the clipboard.

## Package layout

```
frontend/
├── index.html                  HTML shell — Vite injects the JS bundle here
├── vite.config.ts              Vite config — plugins, dev proxy
├── package.json
├── tsconfig*.json              TypeScript project references (app + node)
├── public/                     Static assets served as-is
└── src/
    ├── main.tsx                React root — mounts <App /> into #root
    ├── App.tsx                 Application root — state owner, data fetching
    ├── index.css               Global styles (Tailwind base import)
    ├── types.ts                Shared TypeScript types (LinkItem, URLRecord, ShortenResponse)
    ├── api/
    │   └── client.ts           Typed fetch wrappers for the Go API
    └── components/
        ├── ShortenForm.tsx     URL input form + inline result display
        └── LinksList.tsx       Table of all shortened links for the session
```

## Component hierarchy

```
App
├── ShortenForm          Controlled form; calls shortenURL() on submit
│                        Shows the new short URL inline with a copy button on success
│                        Bubbles the new LinkItem up via onShortened prop
└── LinksList            Renders links[] as a table
    └── CopyButton       (internal) Clipboard button with transient "✓ Copied" feedback
```

## State and data flow

`App` owns the single `links: LinkItem[]` array:

- On mount, `listURLs()` fetches existing links from `GET /api/urls` and populates the list.
  If the request fails, the list starts empty — no error is shown.
- When the user shortens a new URL, `ShortenForm` calls `onShortened(item)`, which prepends
  the new item to `links` so it appears at the top of the table immediately.

There is no client-side router. The app is a single view.

## API client (`src/api/client.ts`)

All backend communication goes through this module. It uses relative paths (`/api/...`) so the
Vite dev proxy handles routing in development — no hardcoded ports in fetch calls.

### `shortenURL(longUrl: string)`

`POST /api/shorten` — submits a long URL and returns the response object from the backend.
The fully-formed `short_url` is built server-side using the backend's `BASE_URL` env var
and returned directly in the response.

On a non-OK response it throws an `Error` with the `error` field from the JSON body, or a
generic message including the HTTP status code. The caller (`ShortenForm`) surfaces this in the UI.

### `listURLs()`

`GET /api/urls` — returns the full list of active URLs, or `null` on failure.

### `buildShortUrl(code: string)`

Constructs the full redirect URL: `VITE_API_BASE_URL + "/" + code`.
The redirect must point directly at the Go server (not through the Vite proxy), so this uses the
env var rather than a relative path.

## Types (`src/types.ts`)

| Type | Usage |
|------|-------|
| `LinkItem` | In-app representation of a link — the shape components work with |
| `URLRecord` | JSON shape returned by `GET /api/urls` — mirrors the Go `URL` struct |
| `ShortenResponse` | JSON shape returned by `POST /api/shorten` — mirrors Go `URLResponse` |

`URLRecord → LinkItem` conversion happens in `App.tsx` (`urlRecordToLinkItem`), keeping
components decoupled from the raw API shape.

## Dev server and proxy

The Vite dev server runs on `http://localhost:5173` by default.

`vite.config.ts` proxies both API calls and short-code redirects to the Go backend:

```ts
proxy: {
  '/api': {
    target: process.env.VITE_API_URL ?? 'http://localhost:8080',
    changeOrigin: true,
  },
  // Forwards /:code to the backend redirect handler.
  // Static assets, Vite internals (/@..., /__...), and the root are served by Vite directly.
  '^/[^/]+$': {
    target: process.env.VITE_API_URL ?? 'http://localhost:8080',
    changeOrigin: true,
    bypass(req) { /* skip Vite-owned paths */ },
  },
}
```

This makes all API calls same-origin from the browser's perspective, so CORS is never
triggered during development. In production the Go backend must send correct CORS headers
(configured via `ALLOWED_ORIGINS` env var on the backend).

To override the backend target locally, set `VITE_API_URL` in `frontend/.env.local`.

## Tech stack

| Package | Version | Purpose |
|---------|---------|---------|
| React | 19 | UI component model |
| TypeScript | ~5.9 | Static typing |
| Vite | latest | Dev server + production bundler |
| Tailwind CSS | v4 | Utility-first styling (via `@tailwindcss/vite` plugin) |
| `babel-plugin-react-compiler` | latest | Enables the React Compiler for automatic memoization |

Tailwind v4 is configured entirely through the Vite plugin — no `tailwind.config.js` needed.

The React Compiler is enabled via the Babel plugin in `vite.config.ts`. It automatically
inserts memoization where needed, replacing manual `useMemo` / `useCallback` calls.

## Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `VITE_API_URL` | No | Origin of the Go backend used by the Vite dev proxy. Defaults to `http://localhost:8080`. Only needed in development — has no effect on production builds. |

All Vite env vars must start with `VITE_` to be accessible in browser code.

## Scripts

| Command | Description |
|---------|-------------|
| `pnpm dev` | Start the Vite dev server with HMR |
| `pnpm build` | Type-check then produce a static bundle in `dist/` |
| `pnpm preview` | Serve the `dist/` build locally |
| `pnpm lint` | Run ESLint across all source files |
