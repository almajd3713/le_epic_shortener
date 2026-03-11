import { useEffect, useState } from 'react';
import { getClicks } from '../api/client';
import type { Click } from '../types';

interface Props {
  shortCode: string;
  shortUrl: string;
  onClose: () => void;
}

// ── helpers ──────────────────────────────────────────────────────────────────

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
}

// Group clicks by calendar date (YYYY-MM-DD in local time)
function groupByDay(clicks: Click[]): { label: string; count: number }[] {
  const map = new Map<string, number>();
  for (const c of clicks) {
    const d = new Date(c.clicked_at);
    const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
    map.set(key, (map.get(key) ?? 0) + 1);
  }
  return Array.from(map.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([label, count]) => ({ label, count }));
}

// Group clicks by hour-of-day (0–23)
function groupByHour(clicks: Click[]): number[] {
  const hours = new Array<number>(24).fill(0);
  for (const c of clicks) {
    hours[new Date(c.clicked_at).getHours()]++;
  }
  return hours;
}

// ── sub-components ────────────────────────────────────────────────────────────

function BarChart({ data }: { data: { label: string; count: number }[] }) {
  if (data.length === 0) return null;
  const max = Math.max(...data.map((d) => d.count), 1);
  const WIDTH = 600;
  const HEIGHT = 120;
  const BAR_GAP = 4;
  const barW = Math.max(4, Math.floor((WIDTH - BAR_GAP * (data.length - 1)) / data.length));

  return (
    <div className="overflow-x-auto">
      <svg
        viewBox={`0 0 ${WIDTH} ${HEIGHT + 24}`}
        className="w-full min-w-[300px]"
        aria-label="Clicks per day"
      >
        {data.map((d, i) => {
          const barH = Math.max(2, Math.round((d.count / max) * HEIGHT));
          const x = i * (barW + BAR_GAP);
          const y = HEIGHT - barH;
          return (
            <g key={d.label}>
              <rect
                x={x}
                y={y}
                width={barW}
                height={barH}
                rx={2}
                className="fill-indigo-500 opacity-80 hover:opacity-100 transition-opacity"
              >
                <title>{d.label}: {d.count} click{d.count !== 1 ? 's' : ''}</title>
              </rect>
              {data.length <= 14 && (
                <text
                  x={x + barW / 2}
                  y={HEIGHT + 16}
                  textAnchor="middle"
                  fontSize={9}
                  className="fill-zinc-500"
                >
                  {d.label.slice(5)} {/* MM-DD */}
                </text>
              )}
            </g>
          );
        })}
      </svg>
    </div>
  );
}

function HourHeatmap({ hours }: { hours: number[] }) {
  const max = Math.max(...hours, 1);
  return (
    <div className="flex gap-1 flex-wrap">
      {hours.map((count, h) => {
        const intensity = count / max;
        // Map 0→zinc-700, 1→indigo-400 via inline opacity
        return (
          <div key={h} className="flex flex-col items-center gap-0.5">
            <div
              title={`${h}:00 — ${count} click${count !== 1 ? 's' : ''}`}
              className="w-7 h-7 rounded flex items-center justify-center text-xs font-mono cursor-default"
              style={{
                backgroundColor: count === 0
                  ? 'rgb(63 63 70)' // zinc-700
                  : `rgba(99 102 241 / ${0.2 + intensity * 0.8})`, // indigo with opacity
                color: intensity > 0.5 ? 'rgb(224 231 255)' : 'rgb(161 163 175)',
              }}
            >
              {count > 0 ? count : ''}
            </div>
            <span className="text-[9px] text-zinc-600 leading-none">{h}</span>
          </div>
        );
      })}
    </div>
  );
}

// ── main component ────────────────────────────────────────────────────────────

export function AnalyticsView({ shortCode, shortUrl, onClose }: Props) {
  const [clicks, setClicks] = useState<Click[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const PAGE_SIZE = 20;

  useEffect(() => {
    let cancelled = false;
    setClicks(null);
    setError(null);
    setPage(1);
    getClicks(shortCode)
      .then((data) => { if (!cancelled) setClicks(data ?? []); })
      .catch((err) => { if (!cancelled) setError(err instanceof Error ? err.message : 'Failed to load'); });
    return () => { cancelled = true; };
  }, [shortCode]);

  const byDay = clicks ? groupByDay(clicks) : [];
  const byHour = clicks ? groupByHour(clicks) : new Array<number>(24).fill(0);

  // Clicks shown newest-first
  const sorted = clicks ? [...clicks].sort(
    (a, b) => new Date(b.clicked_at).getTime() - new Date(a.clicked_at).getTime()
  ) : [];
  const totalPages = Math.ceil(sorted.length / PAGE_SIZE);
  const pageSlice = sorted.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  return (
    // Backdrop
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/70 overflow-y-auto py-8 px-4"
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div className="w-full max-w-3xl bg-zinc-900 rounded-2xl border border-zinc-700 shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-zinc-700">
          <div>
            <h2 className="text-lg font-semibold text-zinc-100">Analytics</h2>
            <a
              href={shortUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-indigo-400 hover:underline font-mono"
            >
              {shortUrl}
            </a>
          </div>
          <button
            onClick={onClose}
            aria-label="Close analytics"
            className="text-zinc-400 hover:text-zinc-100 text-2xl leading-none transition-colors"
          >
            ×
          </button>
        </div>

        <div className="px-6 py-6 space-y-8">
          {/* Loading / error states */}
          {clicks === null && !error && (
            <p className="text-zinc-400 text-sm">Loading analytics…</p>
          )}
          {error && (
            <p className="text-red-400 text-sm bg-red-400/10 border border-red-400/20 rounded-lg px-4 py-2">
              {error}
            </p>
          )}

          {clicks !== null && (
            <>
              {/* Stats row */}
              <div className="grid grid-cols-3 gap-4">
                {[
                  { label: 'Total clicks', value: clicks.length },
                  {
                    label: 'Last click',
                    value: clicks.length
                      ? formatDate(sorted[0].clicked_at)
                      : '—',
                  },
                  {
                    label: 'Busiest day',
                    value: byDay.length
                      ? byDay.reduce((a, b) => (b.count > a.count ? b : a)).label
                      : '—',
                  },
                ].map(({ label, value }) => (
                  <div
                    key={label}
                    className="bg-zinc-800 rounded-xl border border-zinc-700 px-4 py-3"
                  >
                    <p className="text-xs text-zinc-500 uppercase tracking-wide mb-1">{label}</p>
                    <p className="text-xl font-semibold text-zinc-100 truncate">{value}</p>
                  </div>
                ))}
              </div>

              {clicks.length === 0 ? (
                <p className="text-zinc-500 text-sm text-center py-6">
                  No clicks recorded yet for this link.
                </p>
              ) : (
                <>
                  {/* Clicks per day */}
                  <section>
                    <h3 className="text-sm font-medium text-zinc-400 mb-3 uppercase tracking-wide">
                      Clicks per day
                    </h3>
                    <div className="bg-zinc-800 rounded-xl border border-zinc-700 p-4">
                      <BarChart data={byDay} />
                    </div>
                  </section>

                  {/* Hour of day heatmap */}
                  <section>
                    <h3 className="text-sm font-medium text-zinc-400 mb-3 uppercase tracking-wide">
                      Hour of day (local time)
                    </h3>
                    <div className="bg-zinc-800 rounded-xl border border-zinc-700 p-4">
                      <HourHeatmap hours={byHour} />
                    </div>
                  </section>

                  {/* Click log */}
                  <section>
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="text-sm font-medium text-zinc-400 uppercase tracking-wide">
                        Click log
                      </h3>
                      <span className="text-xs text-zinc-600">
                        {sorted.length} total
                      </span>
                    </div>
                    <div className="bg-zinc-800 rounded-xl border border-zinc-700 overflow-hidden">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="text-left text-zinc-500 text-xs uppercase tracking-wider border-b border-zinc-700">
                            <th className="px-4 py-2 font-medium">#</th>
                            <th className="px-4 py-2 font-medium">Clicked at</th>
                          </tr>
                        </thead>
                        <tbody>
                          {pageSlice.map((c, i) => (
                            <tr
                              key={c.id}
                              className="border-b border-zinc-700/50 last:border-0 hover:bg-zinc-700/30 transition-colors"
                            >
                              <td className="px-4 py-2 text-zinc-500 tabular-nums">
                                {(page - 1) * PAGE_SIZE + i + 1}
                              </td>
                              <td className="px-4 py-2 text-zinc-300 tabular-nums">
                                {formatDate(c.clicked_at)}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>

                      {/* Pagination */}
                      {totalPages > 1 && (
                        <div className="flex items-center justify-between px-4 py-3 border-t border-zinc-700 text-xs text-zinc-400">
                          <button
                            onClick={() => setPage((p) => Math.max(1, p - 1))}
                            disabled={page === 1}
                            className="disabled:opacity-30 hover:text-zinc-200 transition-colors"
                          >
                            ← Prev
                          </button>
                          <span>
                            Page {page} / {totalPages}
                          </span>
                          <button
                            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                            disabled={page === totalPages}
                            className="disabled:opacity-30 hover:text-zinc-200 transition-colors"
                          >
                            Next →
                          </button>
                        </div>
                      )}
                    </div>
                  </section>
                </>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
