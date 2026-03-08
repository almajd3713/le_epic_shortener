import type { ShortenResponse, URLRecord } from '../types';

// The Vite dev server proxies /api/* to the Go backend (see vite.config.ts),
// so API calls use relative paths — no CORS issue in development.
// Short URLs are built server-side (using the backend's BASE_URL env var) and returned directly in API responses.


/**
 * POST /api/shorten
 * Resolves with the short code on success.
 * Throws an Error with a human-readable message on failure.
 */
export async function shortenURL(longUrl: string): Promise<ShortenResponse> {
  const expiresAt = new Date(Date.now() + 60 * 60 * 1000).toISOString(); // 1 hour from now
  const res = await fetch('/api/shorten', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ long_url: longUrl, expires_at: expiresAt }),
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
 * Returns null on any non-OK response to avoid crashing the UI.
 */
export async function listURLs(): Promise<URLRecord[] | null> {
  const res = await fetch('/api/urls');
  if (!res.ok) return null;
  return (await res.json()) as URLRecord[];
}

/**
 * GET /geturl/:code
 * Returns the original long URL for a given short code.
 * Throws on non-OK responses.
 */
export async function lookupURL(code: string): Promise<{ long_url: string }> {
  const res = await fetch(`/geturl/${encodeURIComponent(code)}`);
  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `Not found (${res.status})`);
  }
  return (await res.json()) as { long_url: string };
}

/**
 * PATCH /api/toggle/:code
 * Activates or deactivates a shortened URL.
 * Throws on failure.
 */
export async function toggleURL(code: string, action: 'activate' | 'deactivate'): Promise<void> {
  const res = await fetch(`/api/toggle/${encodeURIComponent(code)}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action }),
  });
  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `Request failed (${res.status})`);
  }
}

/**
 * DELETE /api/delete/:code
 * Permanently deletes a shortened URL.
 * Throws on failure.
 */
export async function deleteURL(code: string): Promise<void> {
  const res = await fetch(`/api/delete/${encodeURIComponent(code)}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `Request failed (${res.status})`);
  }
}
