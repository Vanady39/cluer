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