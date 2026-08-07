import { useState } from "react";
import type { TourHint } from'../types/sdk'

export function useHintManager(initialHints: TourHint[] = []) {
  const [hints, setHints] = useState<TourHint[]>(initialHints);

  const addHint = () => {
    setHints((prev) => [
      ...prev,
      {
        id: String(Date.now()),
        step: prev.length + 1,
        title: `Шаг ${prev.length + 1}`,
        content: "",
        selector: "",
        placement: "bottom",
        page_path:"/",
        spotlight: true,
        wait_for_selector: false,
      },
    ]);
  };

  const updateHint = (id: string, field: keyof TourHint, value: string | boolean) => {
    setHints((prev) =>
      prev.map((hint) =>
        hint.id === id ? { ...hint, [field]: value } : hint
      )
    );
  };

  const removeHint = (id: string) => {
    setHints((prev) => prev.filter((hint) => hint.id !== id));
  };

  return { hints, setHints, addHint, updateHint, removeHint };
}