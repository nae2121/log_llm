"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { ArrowLeft, CheckCircle2, CircleX, FileJson2, ShieldAlert } from "lucide-react";
import {
  ConversationDetail,
  Message,
  RiskEvent,
  formatDate,
  getConversationDetail,
} from "@/app/lib/api";
import { RiskBadge, SourceBadge } from "@/app/components/RiskBadge";

export default function ConversationDashboardPage() {
  const params = useParams<{ id: string }>();
  const conversationId = Array.isArray(params.id) ? params.id[0] : params.id;
  const [detail, setDetail] = useState<ConversationDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function load() {
      setLoading(true);
      setError("");
      try {
        setDetail(await getConversationDetail(conversationId));
      } catch (err) {
        setError(err instanceof Error ? err.message : "Conversation request failed");
      } finally {
        setLoading(false);
      }
    }
    if (conversationId) {
      void load();
    }
  }, [conversationId]);

  const comparison = useMemo(() => {
    const rows = new Map<string, { type: string; rule: number; judge: number; maxScore: number }>();
    for (const event of detail?.riskEvents ?? []) {
      if (event.type === "benign") {
        continue;
      }
      const row = rows.get(event.type) ?? {
        type: event.type,
        rule: 0,
        judge: 0,
        maxScore: 0,
      };
      if (event.source.startsWith("rule")) {
        row.rule += 1;
      }
      if (event.source.startsWith("llm_judge")) {
        row.judge += 1;
      }
      row.maxScore = Math.max(row.maxScore, event.score);
      rows.set(event.type, row);
    }
    return Array.from(rows.values()).sort((a, b) => b.maxScore - a.maxScore);
  }, [detail]);

  const eventsByMessage = useMemo(() => {
    const grouped = new Map<string, RiskEvent[]>();
    for (const event of detail?.riskEvents ?? []) {
      if (!event.messageId) {
        continue;
      }
      grouped.set(event.messageId, [...(grouped.get(event.messageId) ?? []), event]);
    }
    return grouped;
  }, [detail]);

  if (loading) {
    return (
      <main className="mx-auto w-full max-w-7xl px-4 py-8 text-sm text-zinc-500 sm:px-6 lg:px-8">
        Loading
      </main>
    );
  }

  if (error || !detail) {
    return (
      <main className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="rounded-lg border border-red-100 bg-red-50 p-4 text-sm text-red-700">
          {error || "Conversation not found"}
        </div>
      </main>
    );
  }

  return (
    <main className="mx-auto w-full max-w-7xl space-y-4 px-4 py-4 sm:px-6 lg:px-8">
      <section className="rounded-lg border border-zinc-200 bg-white p-4">
        <Link
          href="/dashboard"
          className="mb-4 inline-flex items-center gap-2 text-sm font-medium text-zinc-600 hover:text-zinc-950"
        >
          <ArrowLeft className="h-4 w-4" aria-hidden="true" />
          Dashboard
        </Link>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-lg font-semibold text-zinc-950">
              {detail.conversation.title}
            </h1>
            <p className="mt-1 text-xs text-zinc-500">{detail.conversation.id}</p>
          </div>
          <p className="text-sm text-zinc-500">
            Updated {formatDate(detail.conversation.updatedAt)}
          </p>
        </div>
      </section>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="space-y-4">
          {detail.messages.map((message) => (
            <MessagePanel
              key={message.id}
              message={message}
              events={eventsByMessage.get(message.id) ?? []}
            />
          ))}
        </div>

        <aside className="space-y-4">
          <div className="rounded-lg border border-zinc-200 bg-white">
            <div className="flex items-center gap-2 border-b border-zinc-200 p-4">
              <ShieldAlert className="h-4 w-4 text-zinc-500" aria-hidden="true" />
              <h2 className="text-sm font-semibold text-zinc-950">Rule / Judge</h2>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="bg-zinc-50 text-xs uppercase text-zinc-500">
                  <tr>
                    <th className="px-4 py-3 font-semibold">Type</th>
                    <th className="px-4 py-3 font-semibold">Rule</th>
                    <th className="px-4 py-3 font-semibold">Judge</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-100">
                  {comparison.length === 0 ? (
                    <tr>
                      <td colSpan={3} className="px-4 py-6 text-center text-zinc-500">
                        No detections
                      </td>
                    </tr>
                  ) : (
                    comparison.map((row) => (
                      <tr key={row.type}>
                        <td className="px-4 py-3 font-medium text-zinc-900">{row.type}</td>
                        <td className="px-4 py-3 text-zinc-600">{row.rule}</td>
                        <td className="px-4 py-3 text-zinc-600">{row.judge}</td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>

          <div className="rounded-lg border border-zinc-200 bg-white">
            <div className="flex items-center gap-2 border-b border-zinc-200 p-4">
              <FileJson2 className="h-4 w-4 text-zinc-500" aria-hidden="true" />
              <h2 className="text-sm font-semibold text-zinc-950">Detector runs</h2>
            </div>
            <div className="divide-y divide-zinc-100">
              {detail.detectorRuns.length === 0 ? (
                <p className="p-4 text-sm text-zinc-500">No detector runs</p>
              ) : (
                detail.detectorRuns.map((run) => (
                  <details key={run.id} className="group p-4">
                    <summary className="flex cursor-pointer list-none items-start justify-between gap-3">
                      <span>
                        <span className="flex items-center gap-2 text-sm font-semibold text-zinc-900">
                          {run.success ? (
                            <CheckCircle2 className="h-4 w-4 text-emerald-600" aria-hidden="true" />
                          ) : (
                            <CircleX className="h-4 w-4 text-red-600" aria-hidden="true" />
                          )}
                          {run.phase}
                        </span>
                        <span className="mt-1 block text-xs text-zinc-500">
                          {run.detectorModel} · {run.latencyMs}ms
                        </span>
                      </span>
                      <span className="text-xs text-zinc-500">{formatDate(run.createdAt)}</span>
                    </summary>
                    <div className="mt-3 space-y-2">
                      {run.errorMessage ? (
                        <p className="rounded-md bg-red-50 p-2 text-xs text-red-700">
                          {run.errorMessage}
                        </p>
                      ) : null}
                      <pre className="max-h-64 overflow-auto rounded-md bg-zinc-950 p-3 text-xs leading-5 text-zinc-100">
                        {JSON.stringify(run.outputJson, null, 2) || run.rawOutput}
                      </pre>
                    </div>
                  </details>
                ))
              )}
            </div>
          </div>
        </aside>
      </section>
    </main>
  );
}

function MessagePanel({ message, events }: { message: Message; events: RiskEvent[] }) {
  return (
    <article className="rounded-lg border border-zinc-200 bg-white">
      <div className="flex flex-col gap-3 border-b border-zinc-200 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-sm font-semibold text-zinc-950">{message.role}</h2>
          <p className="mt-1 text-xs text-zinc-500">
            {message.model} · {formatDate(message.createdAt)}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {events.slice(0, 3).map((event) => (
            <RiskBadge
              key={`${event.type}-${event.source}-${event.score}`}
              severity={event.severity}
              score={event.score}
            />
          ))}
        </div>
      </div>
      <div className="p-4">
        <p className="whitespace-pre-wrap break-words text-sm leading-6 text-zinc-800">
          <HighlightedText text={message.content} events={events} />
        </p>
      </div>
      {events.length > 0 ? (
        <div className="border-t border-zinc-200 p-4">
          <div className="grid gap-3 md:grid-cols-2">
            {events.map((event) => (
              <div
                key={`${event.id ?? event.type}-${event.source}`}
                className="rounded-lg border border-zinc-200 p-3"
              >
                <div className="flex flex-wrap items-center gap-2">
                  <RiskBadge severity={event.severity} score={event.score} />
                  <SourceBadge source={event.source} />
                </div>
                <p className="mt-2 text-sm font-semibold text-zinc-900">{event.type}</p>
                {event.evidence ? (
                  <p className="mt-2 break-words rounded-md bg-amber-50 px-2 py-1 text-xs text-amber-900">
                    {event.evidence}
                  </p>
                ) : null}
                <p className="mt-2 text-sm leading-6 text-zinc-600">{event.reason}</p>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </article>
  );
}

function HighlightedText({ text, events }: { text: string; events: RiskEvent[] }) {
  const event = events.find((item) => item.evidence && text.includes(item.evidence));
  if (!event?.evidence) {
    return <>{text}</>;
  }

  const parts = text.split(event.evidence);
  return (
    <>
      {parts.map((part, index) => (
        <span key={`${part}-${index}`}>
          {part}
          {index < parts.length - 1 ? (
            <mark className="rounded bg-amber-200 px-1 text-amber-950">
              {event.evidence}
            </mark>
          ) : null}
        </span>
      ))}
    </>
  );
}
