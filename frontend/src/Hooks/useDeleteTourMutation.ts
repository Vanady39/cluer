import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { Tour } from "../types/sdk";
// import { onboardingAPI } from "../Api/onboarding";

export function useDeleteTourMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      // Если API готово — раскомментируйте:
      // await onboardingAPI.deleteTour(id);
      
      // Пока через localStorage
      const tours = JSON.parse(localStorage.getItem("tours") || "[]");
      const updated = tours.filter((item: Tour) => item.id !== id);
      localStorage.setItem("tours", JSON.stringify(updated));
      return updated;
    },
    onSuccess: (updatedScenarios) => {
      queryClient.setQueryData(["tours"], updatedScenarios);
    },
    onError: (error) => {
      console.error("Ошибка удаления:", error);
      alert("Не удалось удалить сценарий");
    },
  });
}