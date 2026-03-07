import type { ShortenResponse, URLRecord } from '../types';

// The Vite dev server proxies /api/* to the Go backend (see vite.config.ts),
// so API calls use relative paths — no CORS issue in development.
// Short URLs are built server-side and returned directly in API responses.


/**
 * POST /api/shorten
 * Resolves with the short code on success.
 * Throws an Error with a human-readable message on failure.
 */
export async function shortenURL(longUrl: string): Promise<ShortenResponse> {
  const res = await fetch('/api/shorten', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ long_url: longUrl }),
  });

  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `Request failed (${res.status})`);
  }

  const data = (await res.json()) as ShortenResponse;
  return data;
}

/**
 * GET /api/urls
 * Returns null when the backend endpoint is not yet implemented (404).
 * See docs/backend-changes-required.md.
 */
export async function listURLs(): Promise<URLRecord[] | null> {
  const res = await fetch('/api/urls');
  if (res.status === 404) return null; // endpoint not yet available
  if (!res.ok) throw new Error(`Failed to load URLs (${res.status})`);
  return (await res.json()) as URLRecord[];
}
