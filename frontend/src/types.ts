// LinkItem is the in-app representation of a shortened URL — used by UI components.
export interface LinkItem {
  shortCode: string;
  longUrl: string;
  shortUrl: string; // fully formed redirect URL, e.g. http://localhost:8080/abc123xy
  createdAt: string; // ISO 8601
}

// ShortenResponse mirrors the JSON body returned by POST /api/shorten.
export interface ShortenResponse {
  shortened_url: string; // the short code only — frontend builds the full URL
}

// URLRecord mirrors the JSON shape that GET /api/urls will return once the
// backend exposes that endpoint. See docs/backend-changes-required.md.
export interface URLRecord {
  short_code: string;
  long_url: string;
  created_at: string;
  expires_at: string | null;
}
