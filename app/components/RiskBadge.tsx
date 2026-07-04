import type { Severity } from "@/app/lib/api";

const severityStyles: Record<Severity, string> = {
  none: "border-zinc-200 bg-zinc-50 text-zinc-600",
  low: "border-emerald-200 bg-emerald-50 text-emerald-700",
  medium: "border-amber-200 bg-amber-50 text-amber-800",
  high: "border-red-200 bg-red-50 text-red-700",
  critical: "border-fuchsia-200 bg-fuchsia-50 text-fuchsia-800",
};

export function RiskBadge({
  severity,
  score,
}: {
  severity: Severity;
  score?: number;
}) {
  return (
    <span
      className={`inline-flex h-7 items-center whitespace-nowrap rounded-md border px-2.5 text-xs font-semibold ${severityStyles[severity]}`}
    >
      {severity}
      {typeof score === "number" ? ` ${score}` : ""}
    </span>
  );
}

export function SourceBadge({ source }: { source: string }) {
  const isJudge = source.startsWith("llm_judge");
  return (
    <span
      className={`inline-flex h-7 items-center whitespace-nowrap rounded-md border px-2.5 text-xs font-medium ${
        isJudge
          ? "border-sky-200 bg-sky-50 text-sky-700"
          : "border-zinc-200 bg-white text-zinc-700"
      }`}
    >
      {source}
    </span>
  );
}
