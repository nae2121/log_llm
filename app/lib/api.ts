export type Severity = "none" | "low" | "medium" | "high" | "critical";

export type Conversation = {
  id: string;
  title: string;
  createdAt: string;
  updatedAt: string;
};

export type Message = {
  id: string;
  conversationId: string;
  role: "user" | "assistant";
  content: string;
  model: string;
  createdAt: string;
};

export type RiskEvent = {
  id?: string;
  conversationId?: string;
  messageId?: string;
  requestId?: string;
  type: string;
  severity: Severity;
  score: number;
  evidence: string;
  reason: string;
  source: string;
  createdAt?: string;
};

export type AggregateRisk = {
  score: number;
  severity: Severity;
  confidence: "low" | "medium" | "high";
  events: RiskEvent[];
};

export type ChatResponse = {
  conversationId: string;
  assistantMessage: Message;
  risk: AggregateRisk;
};

export type DetectorRun = {
  id: string;
  requestId: string;
  detectorName: string;
  detectorModel: string;
  phase: string;
  inputJson: unknown;
  outputJson: unknown;
  rawOutput: string;
  latencyMs: number;
  success: boolean;
  errorMessage: string;
  createdAt: string;
};

export type RiskSummary = {
  severityCounts: Array<{ severity: Severity; count: number }>;
  categoryCounts: Array<{ label: string; count: number }>;
  sourceCounts: Array<{ label: string; count: number }>;
  highRiskConversations: Array<{
    conversationId: string;
    title: string;
    maxScore: number;
    eventCount: number;
    updatedAt: string;
  }>;
};

export type ConversationDetail = {
  conversation: Conversation;
  messages: Message[];
  riskEvents: RiskEvent[];
  detectorRuns: DetectorRun[];
};

const backendURL =
  process.env.NEXT_PUBLIC_BACKEND_URL?.replace(/\/$/, "") ??
  "http://localhost:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${backendURL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    cache: "no-store",
  });

  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new Error(body.error ?? "Request failed");
  }

  return response.json() as Promise<T>;
}

export async function getConversations() {
  const data = await request<{ conversations: Conversation[] }>("/api/conversations");
  return data.conversations;
}

export async function getMessages(conversationId: string) {
  const data = await request<{ messages: Message[] }>(
    `/api/conversations/${conversationId}/messages`,
  );
  return data.messages;
}

export async function sendChat(payload: {
  conversationId: string | null;
  message: string;
  model: string;
}) {
  return request<ChatResponse>("/api/chat", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function getRiskEvents(limit = 100) {
  const data = await request<{ riskEvents: RiskEvent[] }>(
    `/api/dashboard/risk-events?limit=${limit}`,
  );
  return data.riskEvents;
}

export async function getRiskSummary() {
  return request<RiskSummary>("/api/dashboard/risk-summary");
}

export async function getConversationDetail(id: string) {
  return request<ConversationDetail>(`/api/dashboard/conversations/${id}`);
}

export function formatDate(value?: string) {
  if (!value) {
    return "";
  }
  return new Intl.DateTimeFormat("ja-JP", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
