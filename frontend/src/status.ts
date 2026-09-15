import type { FeatureStatus } from "./api";

export const statusLabel: Record<FeatureStatus, string> = {
  open: "Open",
  closed: "Closed",
  whitelisted: "Whitelist only"
};

export const statusDescription: Record<FeatureStatus, string> = {
  open: "Available to every user.",
  closed: "Disabled for every user.",
  whitelisted: "Available only to listed users."
};

export function formatDate(value: string) {
  return new Intl.DateTimeFormat("en", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}
