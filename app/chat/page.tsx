"use client";

import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import {
  AlertTriangle,
  Bot,
  Loader2,
  MessageSquarePlus,
  RefreshCw,
  Send,
  User,
} from "lucide-react";
import {
  AggregateRisk,
  Conversation,
  Message,
  formatDate,
  getConversations,
  getMessages,
  sendChat,
} from "@/app/lib/api";
import { RiskBadge, SourceBadge } from "@/app/components/RiskBadge";

const modelOptions = [
  { value: "ollama", label: "Ollama" },
  { value: "gemini", label: "Gemini" },
  { value: "openai", label: "OpenAI" },
];

export default function ChatPage() {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [latestRisk, setLatestRisk] = useState<AggregateRisk | null>(null);
  const [message, setMessage] = useState("");
  const [model, setModel] = useState("ollama");
  const [loading, setLoading] = useState(false);
  const [loadingConversations, setLoadingConversations] = useState(true);
  const [error, setError] = useState("");
  const bottomRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    async function loadInitialConversations() {
      setLoadingConversations(true);
      try {
        const loaded = await getConversations();
        setConversations(loaded);
        const selectedID = loaded[0]?.id ?? null;
        if (selectedID) {
          setConversationId(selectedID);
          setMessages(await getMessages(selectedID));
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load conversations");
      } finally {
        setLoadingConversations(false);
      }
    }

    void loadInitialConversations();
  }, []);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages.length]);

  const activeConversation = useMemo(
    () => conversations.find((conversation) => conversation.id === conversationId),
    [conversations, conversationId],
  );

  async function refreshConversations(nextID?: string) {
    setLoadingConversations(true);
    try {
      const loaded = await getConversations();
      setConversations(loaded);
      const selectedID = nextID ?? conversationId ?? loaded[0]?.id ?? null;
      if (selectedID) {
        setConversationId(selectedID);
        setMessages(await getMessages(selectedID));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load conversations");
    } finally {
      setLoadingConversations(false);
    }
  }

  async function selectConversation(id: string) {
    setConversationId(id);
    setLatestRisk(null);
    setError("");
    try {
      setMessages(await getMessages(id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load messages");
    }
  }

  function startNewConversation() {
    setConversationId(null);
    setMessages([]);
    setLatestRisk(null);
    setError("");
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const content = message.trim();
    if (!content || loading) {
      return;
    }

    setLoading(true);
    setError("");
    setMessage("");

    const optimistic: Message = {
      id: `local-${Date.now()}`,
      conversationId: conversationId ?? "pending",
      role: "user",
      content,
      model,
      createdAt: new Date().toISOString(),
    };
    setMessages((current) => [...current, optimistic]);

    try {
      const response = await sendChat({
        conversationId,
        message: content,
        model,
      });
      setConversationId(response.conversationId);
      setLatestRisk(response.risk);
      setMessages(await getMessages(response.conversationId));
      await refreshConversations(response.conversationId);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Chat request failed");
      setMessages((current) => current.filter((item) => item.id !== optimistic.id));
      setMessage(content);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="mx-auto grid w-full max-w-7xl gap-4 px-4 py-4 sm:px-6 lg:grid-cols-[280px_minmax(0,1fr)_340px] lg:px-8">
      <aside className="rounded-lg border border-zinc-200 bg-white">
        <div className="flex items-center justify-between border-b border-zinc-200 p-3">
          <h1 className="text-sm font-semibold text-zinc-950">Conversations</h1>
          <button
            type="button"
            onClick={startNewConversation}
            className="flex h-9 w-9 items-center justify-center rounded-md border border-zinc-200 text-zinc-700 hover:bg-zinc-100"
            aria-label="New conversation"
            title="New conversation"
          >
            <MessageSquarePlus className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
        <div className="max-h-[calc(100vh-9rem)] overflow-y-auto p-2">
          {loadingConversations ? (
            <div className="flex items-center gap-2 px-3 py-4 text-sm text-zinc-500">
              <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
              Loading
            </div>
          ) : conversations.length === 0 ? (
            <div className="px-3 py-8 text-sm text-zinc-500">No conversations</div>
          ) : (
            conversations.map((conversation) => (
              <button
                key={conversation.id}
                type="button"
                onClick={() => void selectConversation(conversation.id)}
                className={`mb-1 flex w-full flex-col rounded-md px-3 py-2 text-left transition ${
                  conversation.id === conversationId
                    ? "bg-zinc-950 text-white"
                    : "text-zinc-700 hover:bg-zinc-100"
                }`}
              >
                <span className="line-clamp-2 text-sm font-medium">
                  {conversation.title}
                </span>
                <span
                  className={`mt-1 text-xs ${
                    conversation.id === conversationId ? "text-zinc-300" : "text-zinc-500"
                  }`}
                >
                  {formatDate(conversation.updatedAt)}
                </span>
              </button>
            ))
          )}
        </div>
      </aside>

      <section className="flex min-h-[calc(100vh-7rem)] flex-col rounded-lg border border-zinc-200 bg-white">
        <div className="flex flex-col gap-3 border-b border-zinc-200 p-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 className="text-base font-semibold text-zinc-950">
              {activeConversation?.title ?? "New conversation"}
            </h2>
            <p className="mt-1 text-xs text-zinc-500">
              {conversationId ?? "Unsaved"}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <select
              value={model}
              onChange={(event) => setModel(event.target.value)}
              className="h-10 rounded-md border border-zinc-200 bg-white px-3 text-sm text-zinc-800 outline-none focus:border-zinc-500"
            >
              {modelOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <button
              type="button"
              onClick={() => void refreshConversations(conversationId ?? undefined)}
              className="flex h-10 w-10 items-center justify-center rounded-md border border-zinc-200 text-zinc-700 hover:bg-zinc-100"
              aria-label="Refresh"
              title="Refresh"
            >
              <RefreshCw className="h-4 w-4" aria-hidden="true" />
            </button>
          </div>
        </div>

        <div className="flex-1 space-y-4 overflow-y-auto bg-zinc-50 p-4">
          {messages.length === 0 ? (
            <div className="flex h-full min-h-72 items-center justify-center text-sm text-zinc-500">
              Start a conversation
            </div>
          ) : (
            messages.map((item) => (
              <article
                key={item.id}
                className={`flex gap-3 ${item.role === "user" ? "justify-end" : "justify-start"}`}
              >
                {item.role === "assistant" && (
                  <span className="mt-1 flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-sky-100 text-sky-700">
                    <Bot className="h-4 w-4" aria-hidden="true" />
                  </span>
                )}
                <div
                  className={`max-w-[82%] rounded-lg border px-4 py-3 text-sm leading-6 shadow-sm ${
                    item.role === "user"
                      ? "border-zinc-950 bg-zinc-950 text-white"
                      : "border-zinc-200 bg-white text-zinc-800"
                  }`}
                >
                  <p className="whitespace-pre-wrap break-words">{item.content}</p>
                  <p
                    className={`mt-2 text-xs ${
                      item.role === "user" ? "text-zinc-300" : "text-zinc-500"
                    }`}
                  >
                    {item.role} · {formatDate(item.createdAt)}
                  </p>
                </div>
                {item.role === "user" && (
                  <span className="mt-1 flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-zinc-200 text-zinc-700">
                    <User className="h-4 w-4" aria-hidden="true" />
                  </span>
                )}
              </article>
            ))
          )}
          <div ref={bottomRef} />
        </div>

        {error ? (
          <div className="border-t border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        ) : null}

        <form onSubmit={handleSubmit} className="border-t border-zinc-200 p-4">
          <div className="flex gap-2">
            <textarea
              value={message}
              onChange={(event) => setMessage(event.target.value)}
              rows={2}
              className="min-h-12 flex-1 resize-none rounded-md border border-zinc-200 px-3 py-2 text-sm leading-6 outline-none focus:border-zinc-500"
              placeholder="メッセージ"
            />
            <button
              type="submit"
              disabled={loading || message.trim() === ""}
              className="flex h-12 w-12 shrink-0 items-center justify-center rounded-md bg-zinc-950 text-white transition hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-300"
              aria-label="Send"
              title="Send"
            >
              {loading ? (
                <Loader2 className="h-5 w-5 animate-spin" aria-hidden="true" />
              ) : (
                <Send className="h-5 w-5" aria-hidden="true" />
              )}
            </button>
          </div>
        </form>
      </section>

      <aside className="rounded-lg border border-zinc-200 bg-white">
        <div className="border-b border-zinc-200 p-4">
          <h2 className="flex items-center gap-2 text-sm font-semibold text-zinc-950">
            <AlertTriangle className="h-4 w-4" aria-hidden="true" />
            Latest risk
          </h2>
        </div>
        <div className="space-y-4 p-4">
          {latestRisk ? (
            <>
              <div className="flex items-center justify-between">
                <RiskBadge severity={latestRisk.severity} score={latestRisk.score} />
                <span className="rounded-md border border-zinc-200 px-2.5 py-1 text-xs font-medium text-zinc-600">
                  {latestRisk.confidence}
                </span>
              </div>
              {latestRisk.events.length === 0 ? (
                <p className="text-sm text-zinc-500">No risk events</p>
              ) : (
                <div className="space-y-3">
                  {latestRisk.events.map((event) => (
                    <div
                      key={`${event.type}-${event.source}`}
                      className="rounded-lg border border-zinc-200 p-3"
                    >
                      <div className="flex flex-wrap items-center gap-2">
                        <RiskBadge severity={event.severity} score={event.score} />
                        <SourceBadge source={event.source} />
                      </div>
                      <p className="mt-2 text-sm font-semibold text-zinc-900">
                        {event.type}
                      </p>
                      {event.evidence ? (
                        <p className="mt-2 break-words rounded-md bg-amber-50 px-2 py-1 text-xs text-amber-900">
                          {event.evidence}
                        </p>
                      ) : null}
                      <p className="mt-2 text-sm leading-6 text-zinc-600">
                        {event.reason}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </>
          ) : (
            <p className="text-sm text-zinc-500">No recent result</p>
          )}
        </div>
      </aside>
    </main>
  );
}
