import { useEffect } from "react";

interface Props {
  onSelect?: (selector: string) => void;
}

export function Builder({ onSelect }: Props) {
  useEffect(() => {
    const handleClick = (event: MouseEvent) => {
      event.preventDefault();
      event.stopPropagation();

      const element = event.target as HTMLElement;
      const selector = createSelector(element);

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

function createSelector(element: HTMLElement) {
  if (element.id) {
    return `#${element.id}`;
  }

  if (element.dataset.testid) {
    return `[data-testid="${element.dataset.testid}"]`;
  }

  if (typeof element.className === "string" && element.className.trim()) {
    const className = element.className.split(" ").filter(Boolean).join(".");

    return `${element.tagName.toLowerCase()}.${className}`;
  }

  return element.tagName.toLowerCase();
}
