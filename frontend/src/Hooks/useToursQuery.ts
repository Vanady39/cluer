import { useQuery } from "@tanstack/react-query";
import type { Tour } from "../types/sdk";
// import { onboardingAPI } from "../Api/onboarding";

export function useToursQuery() {
  return useQuery({
    queryKey: ["tours"],
    queryFn: async () => {
      // Если API готово — раскомментируйте:
      // return await onboardingAPI.getAll();
      
      // Пока используем localStorage
      const data = JSON.parse(localStorage.getItem("tours") || "[]");
      return data as Tour[];
    },
  });
}