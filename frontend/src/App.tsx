
import { useState, useEffect } from 'react';
import { ShortenForm } from './components/ShortenForm';
import { LinksList } from './components/LinksList';
import { listURLs, } from './api/client';
import type { LinkItem, URLRecord } from './types';

function urlRecordToLinkItem(record: URLRecord): LinkItem {
  return {
    shortCode: record.short_code,
    longUrl: record.long_url,
    shortUrl: record.short_url,
    createdAt: record.created_at,
  };
}

function App() {
  const [links, setLinks] = useState<LinkItem[]>([]);

  // On mount, attempt to load existing links from the backend.
  // Returns null (404) until GET /api/urls is implemented — silently ignored.
  useEffect(() => {
    listURLs()
      .then((records) => {
        if (records) {
          setLinks(records.map(urlRecordToLinkItem));
        }
      })
      .catch(() => {
        // Backend unavailable or endpoint not yet added — start with empty list.
      });
  }, []);

  const handleShortened = (item: LinkItem) => {
    // Prepend so the most recently created link appears first.
    setLinks((prev) => [item, ...prev]);
  };

  return (
    <div className="bg-zinc-900 min-h-screen text-zinc-100">
      <div className="max-w-3xl mx-auto px-4 py-12">
        <header className="mb-8">
          <h1 className="text-3xl font-bold text-zinc-100">🔗 URL Shortener</h1>
          <p className="text-zinc-400 mt-1">
            Turn long URLs into compact, shareable links
          </p>
        </header>

        <ShortenForm onShortened={handleShortened} />
        <LinksList links={links} />
      </div>
    </div>
  );
}

export default App;
