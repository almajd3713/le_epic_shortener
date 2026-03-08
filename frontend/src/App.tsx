
import { useState } from 'react';
import { ShortenForm } from './components/ShortenForm';
import { LookupForm } from './components/LookupForm';
import { LinksList } from './components/LinksList';
import { listURLs, toggleURL, deleteURL } from './api/client';
import type { LinkItem, URLRecord } from './types';

function urlRecordToLinkItem(record: URLRecord): LinkItem {
  return {
    shortCode: record.short_code,
    longUrl: record.long_url,
    shortUrl: record.short_url,
    createdAt: record.created_at,
    isActive: record.is_active,
    expiresAt: record.expires_at,
  };
}

function App() {
  const [links, setLinks] = useState<LinkItem[]>([]);
  const [loadingAll, setLoadingAll] = useState(false);

  const handleLoadAll = async () => {
    setLoadingAll(true);
    try {
      const records = await listURLs();
      if (records) setLinks(records.map(urlRecordToLinkItem));
    } catch {
      // Backend unavailable — silently ignore
    } finally {
      setLoadingAll(false);
    }
  };

  const handleShortened = (item: LinkItem) => {
    setLinks((prev) => [item, ...prev]);
  };

  const handleToggle = async (code: string, action: 'activate' | 'deactivate') => {
    await toggleURL(code, action);
    setLinks((prev) =>
      prev.map((l) => (l.shortCode === code ? { ...l, isActive: action === 'activate' } : l)),
    );
  };

  const handleDelete = async (code: string) => {
    await deleteURL(code);
    setLinks((prev) => prev.filter((l) => l.shortCode !== code));
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
        <LookupForm />

        <div className="mt-6 flex items-center justify-between">
          <h2 className="text-base font-semibold text-zinc-300">All URLs</h2>
          <button
            onClick={handleLoadAll}
            disabled={loadingAll}
            className="text-sm bg-zinc-700 hover:bg-zinc-600 disabled:opacity-50 disabled:cursor-not-allowed text-zinc-200 px-4 py-1.5 rounded-lg transition-colors"
          >
            {loadingAll ? 'Loading…' : 'Load all URLs'}
          </button>
        </div>

        <LinksList links={links} onToggle={handleToggle} onDelete={handleDelete} />
      </div>
    </div>
  );
}

export default App;
