import type { AgentTraceSummary } from "./api";

export function formatDate(value?: string): string {
  if (!value) return "";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

export function shortPath(path?: string): string {
  if (!path) return "";
  const parts = path.split("/").filter(Boolean);
  if (parts.length <= 4) return path;
  return `.../${parts.slice(-4).join("/")}`;
}

export function commandTitle(session: AgentTraceSummary): string {
  return session.command || session.id;
}

export function statusLabel(status?: string): string {
  return status || "unknown";
}

export function statusClass(status?: string): string {
  return tokenClass(statusLabel(status));
}

export function tagLabel(tag: string): string {
  return tag.replace(/[_-]+/g, " ");
}

export function tagClass(tag: string): string {
  return tokenClass(tag);
}

function tokenClass(value: string): string {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "") || "unknown";
}
