import { useState } from 'react';
import type { LinkItem } from '../types';

/** Renders a copy-to-clipboard button with transient feedback. */
function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard API unavailable
    }
  };

  return (
    <button
      onClick={handleCopy}
      className="text-xs text-zinc-400 hover:text-zinc-200 border border-zinc-600 hover:border-zinc-400 rounded px-2 py-0.5 transition-colors"
    >
      {copied ? '✓ Copied' : 'Copy'}
    </button>
  );
}

interface Props {
  links: LinkItem[];
}

export function LinksList({ links }: Props) {
  if (links.length === 0) return null;

  return (
    <div className="mt-6 bg-zinc-800 rounded-xl border border-zinc-700 overflow-hidden">
      <div className="px-6 py-4 border-b border-zinc-700">
        <h2 className="text-lg font-semibold text-zinc-100">
          Links this session{' '}
          <span className="text-sm font-normal text-zinc-500">({links.length})</span>
        </h2>
        <p className="text-xs text-zinc-500 mt-0.5">
          Persists while the page is open. Once the backend exposes{' '}
          <code className="text-zinc-400">GET /api/urls</code>, this list will load
          from the database on every visit.
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-zinc-500 text-xs uppercase tracking-wider border-b border-zinc-700">
              <th className="px-6 py-3 font-medium">Short URL</th>
              <th className="px-6 py-3 font-medium">Original URL</th>
              <th className="px-6 py-3 font-medium">Created</th>
              <th className="px-6 py-3 font-medium"></th>
            </tr>
          </thead>
          <tbody>
            {links.map((link) => (
              <tr
                key={link.shortCode}
                className="border-b border-zinc-700/50 last:border-b-0 hover:bg-zinc-700/30 transition-colors"
              >
                <td className="px-6 py-3">
                  <a
                    href={link.shortUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-indigo-400 hover:text-indigo-300 font-mono hover:underline"
                  >
                    {link.shortCode}
                  </a>
                </td>
                <td className="px-6 py-3">
                  {/* title reveals the full URL when truncated */}
                  <span
                    className="text-zinc-300 block max-w-xs truncate"
                    title={link.longUrl}
                  >
                    {link.longUrl}
                  </span>
                </td>
                <td className="px-6 py-3 text-zinc-500 whitespace-nowrap">
                  {new Date(link.createdAt).toLocaleTimeString()}
                </td>
                <td className="px-6 py-3">
                  <CopyButton text={link.shortUrl} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
