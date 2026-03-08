import { useState } from 'react';
import type { FormEvent } from 'react';
import { lookupURL } from '../api/client';

export function LookupForm() {
  const [code, setCode] = useState('');
  const [result, setResult] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError(null);
    setResult(null);
    setLoading(true);
    try {
      const data = await lookupURL(code.trim());
      setResult(data.long_url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Not found');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mt-4 bg-zinc-800 rounded-xl p-6 border border-zinc-700">
      <h2 className="text-lg font-semibold mb-4 text-zinc-100">Lookup a short code</h2>

      <form onSubmit={handleSubmit} className="flex gap-2">
        <input
          type="text"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="e.g. abc123xy"
          required
          disabled={loading}
          className="flex-1 bg-zinc-900 border border-zinc-600 rounded-lg px-4 py-2 text-zinc-100 placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent disabled:opacity-50 font-mono"
        />
        <button
          type="submit"
          disabled={loading}
          className="bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium px-5 py-2 rounded-lg transition-colors"
        >
          {loading ? 'Looking up…' : 'Lookup'}
        </button>
      </form>

      {error && (
        <p className="mt-3 text-sm text-red-400 bg-red-400/10 border border-red-400/20 rounded-lg px-4 py-2">
          {error}
        </p>
      )}

      {result && (
        <div className="mt-3 flex items-center gap-3 bg-zinc-900/60 border border-zinc-600 rounded-lg px-4 py-3">
          <span className="text-xs text-zinc-500 whitespace-nowrap">Original URL</span>
          <a
            href={result}
            target="_blank"
            rel="noopener noreferrer"
            className="text-indigo-400 hover:text-indigo-300 hover:underline truncate text-sm"
            title={result}
          >
            {result}
          </a>
        </div>
      )}
    </div>
  );
}
