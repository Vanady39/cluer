import { useQuery } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";

export function useTourLoader(editId: string | null) {
  return useQuery({
    queryKey: ["tour", editId],
    queryFn: async () => {
      if (!editId) return null;
      return await onboardingAPI.getTour(editId);
    },
    enabled: !!editId,
  });
}