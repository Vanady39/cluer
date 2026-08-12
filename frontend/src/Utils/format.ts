export function percent(value: number) {
  return `${Math.round(value * 100)}%`;
}

export function formatDate(value: string) {
  return new Date(value).toLocaleDateString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}

export function formatDateTime(value?: string | null): string {
  if (!value) return "—";
  return new Date(value).toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function getStatusLabel(status: "draft" | "published" | "archived"): string {
  switch (status) {
    case "published":
      return "Опубликована";
    case "draft":
      return "Черновик";
    case "archived":
      return "Архивная";
    default:
      return status;
  }
}