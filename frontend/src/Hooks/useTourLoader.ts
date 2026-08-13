import { useQuery } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { onboardingAPI } from "../Api/onboarding";

export function useTourLoader(editId: string | null) {
  return useQuery({
    queryKey: ["tour", editId],
    queryFn: async () => {
      if (!editId) return null;

      let tour = await onboardingAPI.getTour(editId);
      if (!tour.draft) {
        if (!tour.published) throw new Error("Tour has neither draft nor published version");

        try {
          await onboardingAPI.createDraft(editId);
        } catch (error) {
          if (!isAxiosError(error) || error.response?.status !== 409) throw error;
        }

        tour = await onboardingAPI.getTour(editId);
        if (!tour.draft) throw new Error("Draft was not created");
      }
      return tour;
    },
    enabled: !!editId,
    retry: false,
  });
}
