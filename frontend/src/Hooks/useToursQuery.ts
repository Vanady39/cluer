import { useQuery } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";

export function useToursQuery() {
  return useQuery({
    queryKey: ["tours"],

    queryFn: async () => {
      const tours = await onboardingAPI.getTours();

      const enrichedTours = await Promise.all(
        tours.map(async (tour: any) => {
          try {
            const card = await onboardingAPI.getTour(tour.id);

            return {
              ...tour,
              draft: card.draft,
              published: card.published,
              enabled: card.enabled ?? tour.enabled,
              hints: card.hints ?? [],
            };
          } catch (error) {
            console.error(`Failed to load tour ${tour.id}`, error);

            return {
              ...tour,
              hints: [],
            };
          }
        }),
      );

      return enrichedTours;
    },
  });
}
