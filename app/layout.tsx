import type { Metadata } from "next";
import { AppNav } from "@/app/components/AppNav";
import "./globals.css";

export const metadata: Metadata = {
  title: "LLM Security Monitor",
  description: "Security-monitored LLM chat with risk event dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja" className="h-full antialiased">
      <body className="min-h-full bg-zinc-100 text-zinc-950">
        <AppNav />
        {children}
      </body>
    </html>
  );
}
