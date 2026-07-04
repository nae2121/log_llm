"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { BarChart3, CircleAlert, Database, ListFilter, RefreshCw } from "lucide-react";
import {
  RiskEvent,
  RiskSummary,
  formatDate,
  getRiskEvents,
  getRiskSummary,
} from "@/app/lib/api";
import { RiskBadge, SourceBadge } from "@/app/components/RiskBadge";

export default function DashboardPage() {
  const [summary, setSummary] = useState<RiskSummary | null>(null);
  const [events, setEvents] = useState<RiskEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    void loadDashboard();
  }, []);

  async function loadDashboard() {
    setLoading(true);
    setError("");
    try {
      const [nextSummary, nextEvents] = await Promise.all([
        getRiskSummary(),
        getRiskEvents(100),
      ]);
      setSummary(nextSummary);
      setEvents(nextEvents);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Dashboard request failed");
    } finally {
      setLoading(false);
    }
  }

  const maxCategoryCount = useMemo(
    () => Math.max(1, ...(summary?.categoryCounts.map((item) => item.count) ?? [1])),
    [summary],
  );

  return (
    <main className="mx-auto w-full max-w-7xl space-y-4 px-4 py-4 sm:px-6 lg:px-8">
      <section className="flex flex-col gap-3 rounded-lg border border-zinc-200 bg-white p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-lg font-semibold text-zinc-950">Risk Dashboard</h1>
          <p className="mt-1 text-sm text-zinc-500">
            {loading ? "Loading" : `${events.length} recent events`}
          </p>
        </div>
        <button
          type="button"
          onClick={() => void loadDashboard()}
          className="flex h-10 items-center justify-center gap-2 rounded-md border border-zinc-200 px-3 text-sm font-medium text-zinc-700 hover:bg-zinc-100"
        >
          <RefreshCw className="h-4 w-4" aria-hidden="true" />
          Refresh
        </button>
      </section>

      {error ? (
        <div className="rounded-lg border border-red-100 bg-red-50 p-4 text-sm text-red-700">
          {error}
        </div>
      ) : null}

      <section className="grid gap-4 md:grid-cols-4">
        {(summary?.severityCounts ?? []).map((item) => (
          <div key={item.severity} className="rounded-lg border border-zinc-200 bg-white p-4">
            <div className="flex items-center justify-between">
              <RiskBadge severity={item.severity} />
              <CircleAlert className="h-4 w-4 text-zinc-400" aria-hidden="true" />
            </div>
            <p className="mt-4 text-3xl font-semibold text-zinc-950">{item.count}</p>
          </div>
        ))}
      </section>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="rounded-lg border border-zinc-200 bg-white">
          <div className="flex items-center gap-2 border-b border-zinc-200 p-4">
            <BarChart3 className="h-4 w-4 text-zinc-500" aria-hidden="true" />
            <h2 className="text-sm font-semibold text-zinc-950">Categories</h2>
          </div>
          <div className="space-y-3 p-4">
            {(summary?.categoryCounts ?? []).length === 0 ? (
              <p className="text-sm text-zinc-500">No category counts</p>
            ) : (
              summary?.categoryCounts.map((item) => (
                <div key={item.label} className="grid grid-cols-[180px_minmax(0,1fr)_48px] items-center gap-3">
                  <span className="truncate text-sm font-medium text-zinc-700">
                    {item.label}
                  </span>
                  <div className="h-2 rounded-full bg-zinc-100">
                    <div
                      className="h-2 rounded-full bg-sky-500"
                      style={{ width: `${Math.max(5, (item.count / maxCategoryCount) * 100)}%` }}
                    />
                  </div>
                  <span className="text-right text-sm text-zinc-500">{item.count}</span>
                </div>
              ))
            )}
          </div>
        </div>

        <div className="rounded-lg border border-zinc-200 bg-white">
          <div className="flex items-center gap-2 border-b border-zinc-200 p-4">
            <ListFilter className="h-4 w-4 text-zinc-500" aria-hidden="true" />
            <h2 className="text-sm font-semibold text-zinc-950">Sources</h2>
          </div>
          <div className="space-y-3 p-4">
            {(summary?.sourceCounts ?? []).length === 0 ? (
              <p className="text-sm text-zinc-500">No source counts</p>
            ) : (
              summary?.sourceCounts.map((item) => (
                <div key={item.label} className="flex items-center justify-between gap-3">
                  <SourceBadge source={item.label} />
                  <span className="text-sm font-semibold text-zinc-900">{item.count}</span>
                </div>
              ))
            )}
          </div>
        </div>
      </section>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="rounded-lg border border-zinc-200 bg-white">
          <div className="flex items-center gap-2 border-b border-zinc-200 p-4">
            <Database className="h-4 w-4 text-zinc-500" aria-hidden="true" />
            <h2 className="text-sm font-semibold text-zinc-950">Recent risk events</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full min-w-[760px] text-left text-sm">
              <thead className="bg-zinc-50 text-xs uppercase text-zinc-500">
                <tr>
                  <th className="px-4 py-3 font-semibold">Time</th>
                  <th className="px-4 py-3 font-semibold">Type</th>
                  <th className="px-4 py-3 font-semibold">Severity</th>
                  <th className="px-4 py-3 font-semibold">Source</th>
                  <th className="px-4 py-3 font-semibold">Evidence</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-100">
                {events.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">
                      {loading ? "Loading" : "No risk events"}
                    </td>
                  </tr>
                ) : (
                  events.map((event) => (
                    <tr key={event.id ?? `${event.type}-${event.createdAt}`}>
                      <td className="whitespace-nowrap px-4 py-3 text-zinc-500">
                        {formatDate(event.createdAt)}
                      </td>
                      <td className="px-4 py-3">
                        <Link
                          href={`/dashboard/conversations/${event.conversationId}`}
                          className="font-medium text-zinc-950 hover:underline"
                        >
                          {event.type}
                        </Link>
                      </td>
                      <td className="px-4 py-3">
                        <RiskBadge severity={event.severity} score={event.score} />
                      </td>
                      <td className="px-4 py-3">
                        <SourceBadge source={event.source} />
                      </td>
                      <td className="max-w-sm px-4 py-3">
                        <span className="line-clamp-2 text-zinc-600">
                          {event.evidence || event.reason}
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        <div className="rounded-lg border border-zinc-200 bg-white">
          <div className="border-b border-zinc-200 p-4">
            <h2 className="text-sm font-semibold text-zinc-950">High score conversations</h2>
          </div>
          <div className="divide-y divide-zinc-100">
            {(summary?.highRiskConversations ?? []).length === 0 ? (
              <p className="p-4 text-sm text-zinc-500">No high score conversations</p>
            ) : (
              summary?.highRiskConversations.map((conversation) => (
                <Link
                  key={conversation.conversationId}
                  href={`/dashboard/conversations/${conversation.conversationId}`}
                  className="block p-4 hover:bg-zinc-50"
                >
                  <div className="flex items-start justify-between gap-3">
                    <p className="line-clamp-2 text-sm font-semibold text-zinc-900">
                      {conversation.title}
                    </p>
                    <span className="rounded-md bg-zinc-950 px-2 py-1 text-xs font-semibold text-white">
                      {conversation.maxScore}
                    </span>
                  </div>
                  <p className="mt-2 text-xs text-zinc-500">
                    {conversation.eventCount} events · {formatDate(conversation.updatedAt)}
                  </p>
                </Link>
              ))
            )}
          </div>
        </div>
      </section>
    </main>
  );
}
