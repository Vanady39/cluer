import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import type { Tour, TourHint } from "../types/sdk";

export function useSaveScenario(editId: string | null) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      title,
      description,
      hints,
      status,
    }: {
      title: string;
      description: string;
      hints: TourHint[];
      status: "draft" | "published";
    }) => {
      const scenario = {
        id: editId || String(Date.now()),
        title,
        description,
        status,
        target_path: "/",
        priority: 1,
        trigger_type: "on_load",
        audience: {
          show_once: true,
          max_shows: 1,
          only_new: false,
        },
        hints,
        updated_at: new Date().toISOString(),
      };

      const tours = JSON.parse(localStorage.getItem("tours") || "[]");
      const updatedTours = editId
        ? tours.map((tour: Tour) => tour.id === editId ? scenario : tour)
        : [...tours, scenario];

      localStorage.setItem("tours", JSON.stringify(updatedTours));
      return scenario.id;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tours"] });
      navigate("/admin/scenarios");
    },
    onError: (error) => {
      console.error("Ошибка сохранения:", error);
      alert("Не удалось сохранить сценарий");
    },
  });
}