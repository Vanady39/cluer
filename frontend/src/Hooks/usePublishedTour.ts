import { useQuery } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";

export function usePublishedTour(path: string) {
  return useQuery({
    queryKey: ["published-tour", path],
    queryFn: () => onboardingAPI.getPublished(path),
    select: (tours) => tours[0],
  });
}