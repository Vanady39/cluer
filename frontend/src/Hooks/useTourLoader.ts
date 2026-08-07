import { useQuery } from "@tanstack/react-query";
import type { Tour } from "../types/sdk";

export function useTourLoader(editId: string | null) {
  return useQuery({
    queryKey: ["tour", editId],
    queryFn: async () => {
      if (!editId) return null;
      const tours = JSON.parse(localStorage.getItem("tours") || "[]");
      return tours.find((item: Tour) => item.id === editId) || null;
    },
    enabled: !!editId,
  });
}