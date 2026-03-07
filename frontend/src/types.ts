// LinkItem is the in-app representation of a shortened URL — used by UI components.
export interface LinkItem {
  shortCode: string;
  longUrl: string;
  shortUrl: string; // fully formed redirect URL, e.g. http://localhost:8080/abc123xy
  createdAt: string; // ISO 8601
}

// ShortenResponse mirrors the JSON body returned by POST /api/shorten.
export interface ShortenResponse {
  short_code: string;
  short_url: string;
  created_at: string; // ISO 8601
}

// URLRecord mirrors the JSON shape returned by GET /api/urls.
export interface URLRecord {
  short_code: string;
  long_url: string;
  short_url: string;  // fully formed redirect URL, built server-side
  created_at: string;
  expires_at: string | null;
}
