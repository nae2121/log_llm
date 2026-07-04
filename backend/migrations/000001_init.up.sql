CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS conversations (
  id UUID PRIMARY KEY,
  title TEXT NOT NULL DEFAULT 'New conversation',
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS messages (
  id UUID PRIMARY KEY,
  conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  role TEXT NOT NULL,
  content TEXT NOT NULL,
  model TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_created ON messages(conversation_id, created_at);

CREATE TABLE IF NOT EXISTS llm_requests (
  id UUID PRIMARY KEY,
  conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  user_message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  assistant_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
  system_prompt TEXT NOT NULL,
  user_input TEXT NOT NULL,
  model_output TEXT NOT NULL DEFAULT '',
  chat_model TEXT NOT NULL,
  latency_ms INTEGER NOT NULL DEFAULT 0,
  prompt_tokens INTEGER NOT NULL DEFAULT 0,
  completion_tokens INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_llm_requests_conversation_created ON llm_requests(conversation_id, created_at);

CREATE TABLE IF NOT EXISTS risk_events (
  id UUID PRIMARY KEY,
  conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
  request_id UUID NOT NULL REFERENCES llm_requests(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  severity TEXT NOT NULL,
  score INTEGER NOT NULL CHECK (score >= 0 AND score <= 100),
  evidence TEXT NOT NULL DEFAULT '',
  reason TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_risk_events_conversation_created ON risk_events(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_events_score ON risk_events(score DESC);
CREATE INDEX IF NOT EXISTS idx_risk_events_type ON risk_events(type);
CREATE INDEX IF NOT EXISTS idx_risk_events_source ON risk_events(source);

CREATE TABLE IF NOT EXISTS detector_runs (
  id UUID PRIMARY KEY,
  request_id UUID NOT NULL REFERENCES llm_requests(id) ON DELETE CASCADE,
  detector_name TEXT NOT NULL,
  detector_model TEXT NOT NULL,
  phase TEXT NOT NULL,
  input_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  output_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  raw_output TEXT NOT NULL DEFAULT '',
  latency_ms INTEGER NOT NULL DEFAULT 0,
  success BOOLEAN NOT NULL DEFAULT false,
  error_message TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_detector_runs_request_created ON detector_runs(request_id, created_at DESC);
