import { useTourRunner } from "../../Hooks/useTourRunner";
import type { Tour } from "../../types/tour";
import { Hint } from "./Hint";

interface Props {
  tour: Tour;
  onClose: () => void;
  isPreview?: boolean;
}

export function TourRunner({ tour, onClose, isPreview = false }: Props) {
  const { step, element, hint, hints, isHintPage, next, skip, hasHint } = useTourRunner(tour, isPreview, onClose);

  if (!hint || !isHintPage) return null;
  if (hint.placement !== "center" && !element) return null;
  if (!hasHint || !isHintPage) return null;

  return (
    <Hint
      hint={hint}
      element={element}
      step={step}
      total={hints.length}
      next={next}
      skip={skip}
    />
  );
}