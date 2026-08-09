import { useQuery } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";

export function useTourAnalytics(
  tourId?: string,
) {
  return useQuery({
    queryKey: [
      "tour-analytics",
      tourId,
    ],

    queryFn: () => {
      if (!tourId) {
        throw new Error(
          "Tour ID is required",
        );
      }

      return onboardingAPI.getAnalytics(
        tourId,
      );
    },

    enabled: Boolean(tourId),
  });
}