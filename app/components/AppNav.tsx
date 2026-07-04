"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Activity, MessageSquareText, ShieldCheck } from "lucide-react";

const items = [
  { href: "/chat", label: "Chat", icon: MessageSquareText },
  { href: "/dashboard", label: "Dashboard", icon: Activity },
];

export function AppNav() {
  const pathname = usePathname();

  return (
    <header className="border-b border-zinc-200 bg-white">
      <div className="mx-auto flex min-h-16 w-full max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <Link href="/chat" className="flex items-center gap-3 text-zinc-950">
          <span className="flex h-9 w-9 items-center justify-center rounded-md bg-zinc-950 text-white">
            <ShieldCheck className="h-5 w-5" aria-hidden="true" />
          </span>
          <span className="text-sm font-semibold tracking-normal">
            LLM Security Monitor
          </span>
        </Link>
        <nav className="flex items-center gap-1">
          {items.map((item) => {
            const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
            const Icon = item.icon;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex h-10 items-center gap-2 rounded-md px-3 text-sm font-medium transition ${
                  active
                    ? "bg-zinc-950 text-white"
                    : "text-zinc-600 hover:bg-zinc-100 hover:text-zinc-950"
                }`}
              >
                <Icon className="h-4 w-4" aria-hidden="true" />
                <span>{item.label}</span>
              </Link>
            );
          })}
        </nav>
      </div>
    </header>
  );
}
