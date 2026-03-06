import { useState } from 'react';
import type { FormEvent } from 'react';
import { shortenURL, } from '../api/client';
import type { LinkItem } from '../types';

interface Props {
  onShortened: (item: LinkItem) => void;
}

export function ShortenForm({ onShortened }: Props) {
  const [url, setUrl] = useState('');
  const [result, setResult] = useState<LinkItem | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError(null);
    setResult(null);
    setLoading(true);

    try {
      const { short_code, short_url, created_at } = await shortenURL(url);
      const item: LinkItem = {
        shortCode: short_code,
        longUrl: url,
        shortUrl: short_url,
        createdAt: created_at
      };
      setResult(item);
      onShortened(item);
      setUrl('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    if (!result) return;
    try {
      await navigator.clipboard.writeText(result.shortUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard API unavailable (non-HTTPS or blocked by browser)
    }
  };

  return (
    <div className="bg-zinc-800 rounded-xl p-6 border border-zinc-700">
      <h2 className="text-lg font-semibold mb-4 text-zinc-100">Shorten a URL</h2>

      {/* The form submits via React's onSubmit — not a native page reload */}
      <form onSubmit={handleSubmit} className="flex gap-2">
        <input
          type="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://example.com/very/long/url"
          required
          disabled={loading}
          className="flex-1 bg-zinc-900 border border-zinc-600 rounded-lg px-4 py-2 text-zinc-100 placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent disabled:opacity-50"
        />
        <button
          type="submit"
          disabled={loading}
          className="bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium px-5 py-2 rounded-lg transition-colors"
        >
          {loading ? 'Shortening…' : 'Shorten'}
        </button>
      </form>

      {/* Error response from the API is surfaced here (e.g. invalid URL, 500) */}
      {error && (
        <p className="mt-3 text-sm text-red-400 bg-red-400/10 border border-red-400/20 rounded-lg px-4 py-2">
          {error}
        </p>
      )}

      {result && (
        <div className="mt-3 flex items-center gap-3 bg-green-500/10 border border-green-500/20 rounded-lg px-4 py-2">
          <span className="text-green-500 text-lg leading-none">✓</span>
          {/* Opens the redirect URL in a new tab — target="_blank" requires rel="noopener" */}
          <a
            href={result.shortUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="flex-1 text-green-400 font-mono text-sm hover:underline truncate"
          >
            {result.shortUrl}
          </a>
          <button
            onClick={handleCopy}
            className="text-zinc-300 hover:text-white text-sm font-medium transition-colors shrink-0"
          >
            {copied ? '✓ Copied!' : 'Copy'}
          </button>
        </div>
      )}
    </div>
  );
}
