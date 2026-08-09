import { useState } from "react";
import type { TourHint } from "../types/sdk";

export function useHintManager(
  initialHints: TourHint[] = [],
) {
  const [hints, setHints] =
    useState<TourHint[]>(initialHints);

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

        media_url: "",
        spotlight: true,

        required: false,
        wait_for_selector: false,

        input_placeholder: "",
        expected_input: "",
      },
    ]);
  };

  const updateHint = <
    K extends keyof TourHint,
  >(
    id: string,
    field: K,
    value: TourHint[K],
  ) => {
    setHints((prev) =>
      prev.map((hint) =>
        hint.id === id
          ? {
              ...hint,
              [field]: value,
            }
          : hint,
      ),
    );
  };

  const removeHint = (
    id: string,
  ) => {
    setHints((prev) =>
      prev
        .filter(
          (hint) => hint.id !== id,
        )
        .map((hint, index) => ({
          ...hint,
          step: index + 1,
        })),
    );
  };

  return {
    hints,
    setHints,
    addHint,
    updateHint,
    removeHint,
  };
}