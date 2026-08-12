import { useQuery } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";
import type { TourAnalyticsItem } from "../types/sdk";
import { getCurrentApp } from "../Api/Helpers/Helpers";

export function useAllToursAnalytics() {
  return useQuery({
    queryKey: ["all-tours-analytics"],

    queryFn: async (): Promise<TourAnalyticsItem[]> => {
      const app = await getCurrentApp();
      const tours = await onboardingAPI.getTours(app.id);

      return Promise.all(
        tours.map(async (tour) => {
          try {
            const card = await onboardingAPI.getTour(tour.id);
            if (!card.published) {
              return {
                tour: {
                  id: tour.id,
                  title: tour.title || "Без названия",
                  description: tour.description,
                },
                analytics: null,
                unpublished: true,
              };
            }
            const analytics = await onboardingAPI.getAnalytics(tour.id);
            return {
              tour: {
                id: tour.id,
                title: tour.title || "Без названия",
                description: tour.description,
              },
              analytics,
            };
          } catch (error) {
            console.error(`[Analytics] Failed for tour ${tour.id}`, error);

            return {
              tour: {
                id: tour.id,
                title: tour.title || "Без названия",
                description: tour.description,
              },
              analytics: null,

              error:
                error instanceof Error
                  ? error.message
                  : "Не удалось загрузить аналитику",
            };
          }
        }),
      );
    },
  });
}
