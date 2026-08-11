import { useQuery } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";
import type { TourListItem } from "../types/sdk";

export function useToursQuery() {
  return useQuery({
    queryKey: ["tours"],
    queryFn: async () => {
      const tours = await onboardingAPI.getTours();

      const enrichedTours = await Promise.all(
        tours.map(async (tour: TourListItem) => {
          try {
            const card = await onboardingAPI.getTour(tour.id);
            return {
              ...tour,
              draft: card.draft,
              published: card.published,
              enabled: card.enabled,
              hints: card.hints ?? [],
            };
          } catch (error: unknown) {
            console.error(`Failed to load tour ${tour.id}`, error);
            return {
              ...tour,
              draft: null,
              published: null,
              enabled: false,
              hints: [],
            };
          }
        }),
      );

      return enrichedTours;
    },
  });
}