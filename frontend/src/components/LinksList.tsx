import { useState } from 'react';
import type { LinkItem } from '../types';

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
  onToggle: (code: string, action: 'activate' | 'deactivate') => Promise<void>;
  onDelete: (code: string) => Promise<void>;
}

export function LinksList({ links, onToggle, onDelete }: Props) {
  const [busy, setBusy] = useState<Record<string, boolean>>({});

  if (links.length === 0) return null;

  const withBusy = (code: string, fn: () => Promise<void>) => async () => {
    setBusy((prev) => ({ ...prev, [code]: true }));
    try {
      await fn();
    } finally {
      setBusy((prev) => ({ ...prev, [code]: false }));
    }
  };

  return (
    <div className="mt-6 bg-zinc-800 rounded-xl border border-zinc-700 overflow-hidden">
      <div className="px-6 py-4 border-b border-zinc-700">
        <h2 className="text-lg font-semibold text-zinc-100">
          URLs{' '}
          <span className="text-sm font-normal text-zinc-500">({links.length})</span>
        </h2>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-zinc-500 text-xs uppercase tracking-wider border-b border-zinc-700">
              <th className="px-6 py-3 font-medium">Status</th>
              <th className="px-6 py-3 font-medium">Short URL</th>
              <th className="px-6 py-3 font-medium">Original URL</th>
              <th className="px-6 py-3 font-medium">Created</th>
              <th className="px-6 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {links.map((link) => {
              const isBusy = busy[link.shortCode] ?? false;
              return (
                <tr
                  key={link.shortCode}
                  className="border-b border-zinc-700/50 last:border-b-0 hover:bg-zinc-700/30 transition-colors"
                >
                  <td className="px-6 py-3">
                    <span
                      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        link.isActive
                          ? 'bg-emerald-500/15 text-emerald-400'
                          : 'bg-zinc-600/40 text-zinc-400'
                      }`}
                    >
                      {link.isActive ? 'Active' : 'Inactive'}
                    </span>
                  </td>
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
                    <span
                      className="text-zinc-300 block max-w-xs truncate"
                      title={link.longUrl}
                    >
                      {link.longUrl}
                    </span>
                  </td>
                  <td className="px-6 py-3 text-zinc-500 whitespace-nowrap">
                    {new Date(link.createdAt).toLocaleString()}
                  </td>
                  <td className="px-6 py-3">
                    <div className="flex items-center gap-2">
                      <CopyButton text={link.shortUrl} />
                      <button
                        disabled={isBusy}
                        onClick={withBusy(
                          link.shortCode,
                          () => onToggle(link.shortCode, link.isActive ? 'deactivate' : 'activate'),
                        )}
                        className="text-xs border rounded px-2 py-0.5 transition-colors disabled:opacity-40 disabled:cursor-not-allowed text-zinc-400 hover:text-zinc-200 border-zinc-600 hover:border-zinc-400"
                      >
                        {isBusy ? '…' : link.isActive ? 'Deactivate' : 'Activate'}
                      </button>
                      <button
                        disabled={isBusy}
                        onClick={withBusy(link.shortCode, () => onDelete(link.shortCode))}
                        className="text-xs border rounded px-2 py-0.5 transition-colors disabled:opacity-40 disabled:cursor-not-allowed text-red-500 hover:text-red-400 border-red-800 hover:border-red-600"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
