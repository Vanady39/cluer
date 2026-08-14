import { useEffect } from "react";
import { createSelector } from "./selector";

interface Props {
  onSelect?: (selector: string) => void;
}

export function Builder({ onSelect }: Props) {
  useEffect(() => {
    const handleClick = (event: MouseEvent) => {
      event.preventDefault();
      event.stopPropagation();

      const selector = createSelector(event.target as HTMLElement);
      if (window.opener) {
        window.opener.postMessage(
          {
            type: "SELECTOR_SELECTED",
            selector,
          },
          "*",
        );
        window.close();
        return;
      }
      onSelect?.(selector);
    };
    document.addEventListener("click", handleClick, true);
    return () => {
      document.removeEventListener("click", handleClick, true);
    };
  }, [onSelect]);
  return null;
}
